package parser

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"strconv"

	"golang.org/x/tools/go/packages"
)

// GetterDef describes a getter
type GetterDef struct {
	ID      int
	Plural  int
	Context int
	Domain  int
}

const gotextPackagePath = "github.com/leonelquinteros/gotext"

// MaxArgIndex returns the largest argument index
func (d *GetterDef) MaxArgIndex() int {
	return max(d.ID, d.Plural, d.Context, d.Domain)
}

// list of supported getter
var gotextGetter = map[string]GetterDef{
	"Get":    {0, -1, -1, -1},
	"GetN":   {0, 1, -1, -1},
	"GetD":   {1, -1, -1, 0},
	"GetND":  {1, 2, -1, 0},
	"GetC":   {0, -1, 1, -1},
	"GetNC":  {0, 1, 3, -1},
	"GetDC":  {1, -1, 2, 0},
	"GetNDC": {1, 2, 4, 0},
}

// GoFile handles the parsing of one go file
type GoFile struct {
	FilePath string
	BasePath string
	Data     *DomainMap

	FileSet *token.FileSet
	PkgConf *packages.Config

	ImportedPackages map[string]*packages.Package
}

// GetType from ident object
func (g *GoFile) GetType(ident *ast.Ident) types.Object {
	return g.getObject(ident)
}

// getExprType from any expression
func (g *GoFile) getExprType(expr ast.Expr) types.Type {
	if g == nil || expr == nil {
		return nil
	}

	for _, pkg := range g.ImportedPackages {
		if pkg == nil || pkg.TypesInfo == nil {
			continue
		}
		if tv, ok := pkg.TypesInfo.Types[expr]; ok {
			return tv.Type
		}
	}

	if ident, ok := expr.(*ast.Ident); ok {
		if object := g.getObject(ident); object != nil {
			return object.Type()
		}
	}
	if paren, ok := expr.(*ast.ParenExpr); ok {
		return g.getExprType(paren.X)
	}
	return nil
}

// CheckType for gotext object
func (g *GoFile) CheckType(rawType types.Type) bool {
	if rawType == nil {
		return false
	}

	switch t := rawType.(type) {
	case *types.Pointer:
		return g.CheckType(t.Elem())

	case *types.Named:
		object := t.Obj()
		return object != nil && object.Pkg() != nil && object.Pkg().Path() == gotextPackagePath

	case *types.Alias:
		return g.CheckType(t.Rhs())

	case *types.Interface:
		if t.NumMethods() == 0 {
			return false
		}
		for idx := range t.NumMethods() {
			method := t.Method(idx)
			if method.Pkg() == nil || method.Pkg().Path() != gotextPackagePath {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// InspectCallExpr inspects the call expression
func (g *GoFile) InspectCallExpr(n *ast.CallExpr) {
	if g == nil || n == nil {
		return
	}

	var name string
	var object types.Object
	switch fun := n.Fun.(type) {
	case *ast.Ident:
		name = fun.Name
		if _, ok := gotextGetter[name]; !ok {
			return
		}
		object = g.getObject(fun)
		if object != nil {
			if !isGotextGetterObject(object, name) {
				return
			}
		} else if !g.isGotextDotImport(fun) {
			return
		}

	case *ast.SelectorExpr:
		if fun.Sel == nil {
			return
		}
		name = fun.Sel.Name
		if _, ok := gotextGetter[name]; !ok {
			return
		}
		object = g.selectorObject(fun)
		if object != nil {
			if !isGotextGetterObject(object, name) {
				return
			}
			if signature, ok := object.Type().(*types.Signature); ok && signature.Recv() != nil {
				if receiverType := g.getExprType(fun.X); receiverType != nil && !g.CheckType(receiverType) {
					return
				}
			}
		} else if receiverType := g.getExprType(fun.X); receiverType != nil {
			if !g.CheckType(receiverType) {
				return
			}
		} else if !g.isGotextPackageSelector(fun) {
			return
		}

	default:
		return
	}

	args := make([]*ast.BasicLit, len(n.Args))
	resolving := make(map[types.Object]bool)
	for idx, arg := range n.Args {
		args[idx] = g.resolveStringLiteral(arg, n.Pos(), resolving)
	}

	g.ParseGetter(gotextGetter[name], args, g.callPosition(n))
}

func isGotextGetterObject(object types.Object, name string) bool {
	function, ok := object.(*types.Func)
	return ok && function.Name() == name && function.Pkg() != nil && function.Pkg().Path() == gotextPackagePath
}

func (g *GoFile) selectorObject(expr *ast.SelectorExpr) types.Object {
	if g == nil || expr == nil || expr.Sel == nil {
		return nil
	}
	for _, pkg := range g.ImportedPackages {
		if pkg == nil || pkg.TypesInfo == nil {
			continue
		}
		if selection := pkg.TypesInfo.Selections[expr]; selection != nil && selection.Obj() != nil {
			return selection.Obj()
		}
		if object := pkg.TypesInfo.Uses[expr.Sel]; object != nil {
			return object
		}
	}
	return nil
}

func (g *GoFile) isGotextPackageSelector(expr *ast.SelectorExpr) bool {
	if g == nil || expr == nil {
		return false
	}
	ident, ok := expr.X.(*ast.Ident)
	if !ok || ident.Obj != nil {
		return false
	}
	pkg, ok := g.ImportedPackages[ident.Name]
	return ok && pkg != nil && pkg.PkgPath == gotextPackagePath
}

func (g *GoFile) isGotextDotImport(ident *ast.Ident) bool {
	if g == nil || ident == nil || ident.Obj != nil {
		return false
	}
	pkg, ok := g.ImportedPackages["."]
	return ok && pkg != nil && pkg.PkgPath == gotextPackagePath
}

func (g *GoFile) callPosition(n *ast.CallExpr) string {
	if g == nil || n == nil {
		return ""
	}
	path := g.FilePath
	if g.BasePath != "" && path != "" {
		if relative, err := filepath.Rel(g.BasePath, path); err == nil {
			path = relative
		}
	}
	line := 0
	if g.FileSet != nil {
		line = g.FileSet.Position(n.Lparen).Line
	}
	return fmt.Sprintf("%s:%d", path, line)
}

// ParseGetter parses the getter function
func (g *GoFile) ParseGetter(def GetterDef, args []*ast.BasicLit, pos string) {
	if g == nil || g.Data == nil {
		return
	}

	// check if enough arguments are given
	if len(args) <= def.MaxArgIndex() {
		return
	}

	// get domain
	var domain string
	if def.Domain != -1 {
		var ok bool
		domain, ok = getStringArgument(args, def.Domain, "domain", pos)
		if !ok {
			return
		}
	}

	// only handle function calls with strings as ID
	msgID, ok := getStringArgument(args, def.ID, "ID", pos)
	if !ok {
		return
	}

	trans := Translation{
		MsgID:           msgID,
		SourceLocations: []string{pos},
	}
	if def.Plural != -1 {
		// plural ID must be a string
		msgIDPlural, ok := getStringArgument(args, def.Plural, "plural", pos)
		if !ok {
			return
		}
		trans.MsgIDPlural = msgIDPlural
	}
	if def.Context != -1 {
		// Context must be a string
		if !isStringArgument(args, def.Context, "context", pos) {
			return
		}
		trans.Context = args[def.Context].Value
	}

	g.Data.AddTranslation(domain, &trans)
}

func getStringArgument(args []*ast.BasicLit, index int, name, pos string) (string, bool) {
	if !isStringArgument(args, index, name, pos) {
		return "", false
	}

	value, err := strconv.Unquote(args[index].Value)
	if err != nil {
		log.Printf("ERR: Unsupported call at %s (%s is not a valid string)", pos, name)
		return "", false
	}
	return value, true
}

func isStringArgument(args []*ast.BasicLit, index int, name, pos string) bool {
	if index < 0 || index >= len(args) || args[index] == nil || args[index].Kind != token.STRING {
		log.Printf("ERR: Unsupported call at %s (%s is not a string)", pos, name)
		return false
	}
	return true
}

func (g *GoFile) resolveStringLiteral(expr ast.Expr, before token.Pos, resolving map[types.Object]bool) *ast.BasicLit {
	if g == nil || expr == nil {
		return nil
	}
	if resolving == nil {
		resolving = make(map[types.Object]bool)
	}
	return g.resolveStringLiteralState(expr, before, resolving, make(map[ast.Expr]bool))
}

func (g *GoFile) resolveStringLiteralState(
	expr ast.Expr,
	before token.Pos,
	resolving map[types.Object]bool,
	resolvingExpr map[ast.Expr]bool,
) *ast.BasicLit {
	if g == nil || expr == nil {
		return nil
	}

	switch literal := expr.(type) {
	case *ast.BasicLit:
		if literal != nil && literal.Kind == token.STRING {
			return literal
		}
		return nil
	}

	if literal := g.constantStringLiteral(expr); literal != nil {
		return literal
	}
	if resolvingExpr[expr] {
		return nil
	}
	resolvingExpr[expr] = true
	defer delete(resolvingExpr, expr)

	switch value := expr.(type) {
	case *ast.ParenExpr:
		if value == nil {
			return nil
		}
		return g.resolveStringLiteralState(value.X, before, resolving, resolvingExpr)

	case *ast.BinaryExpr:
		if value == nil || value.Op != token.ADD {
			return nil
		}
		left := g.resolveStringLiteralState(value.X, before, resolving, resolvingExpr)
		right := g.resolveStringLiteralState(value.Y, before, resolving, resolvingExpr)
		if left == nil || right == nil {
			return nil
		}
		leftValue, leftErr := strconv.Unquote(left.Value)
		rightValue, rightErr := strconv.Unquote(right.Value)
		if leftErr != nil || rightErr != nil {
			return nil
		}
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(leftValue + rightValue),
		}

	case *ast.Ident:
		if value == nil {
			return nil
		}
		object := g.getObject(value)
		if object != nil {
			if resolving[object] || g.isMutatedBefore(object, before) {
				return nil
			}
			if constant, ok := object.(*types.Const); ok {
				return literalFromConstantValue(constant.Val())
			}
			resolving[object] = true
			defer delete(resolving, object)
		}

		if value.Obj != nil {
			switch declaration := value.Obj.Decl.(type) {
			case *ast.ValueSpec:
				for idx, name := range declaration.Names {
					if name.Name == value.Name && idx < len(declaration.Values) {
						return g.resolveStringLiteralState(
							declaration.Values[idx],
							declaration.Pos(),
							resolving,
							resolvingExpr,
						)
					}
				}
			case *ast.AssignStmt:
				for idx, lhs := range declaration.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && id.Name == value.Name && idx < len(declaration.Rhs) {
						return g.resolveStringLiteralState(
							declaration.Rhs[idx],
							declaration.Pos(),
							resolving,
							resolvingExpr,
						)
					}
				}
			}
		}
		if object != nil {
			if declaration, declarationPos, ok := g.declarationForObject(object); ok {
				return g.resolveStringLiteralState(declaration, declarationPos, resolving, resolvingExpr)
			}
		}
	}
	return nil
}

func literalFromConstantValue(value constant.Value) *ast.BasicLit {
	if value == nil || value.Kind() != constant.String {
		return nil
	}
	return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(constant.StringVal(value))}
}

func (g *GoFile) constantStringLiteral(expr ast.Expr) *ast.BasicLit {
	if g == nil || expr == nil {
		return nil
	}
	for _, pkg := range g.ImportedPackages {
		if pkg == nil || pkg.TypesInfo == nil {
			continue
		}
		if value, ok := pkg.TypesInfo.Types[expr]; ok {
			if literal := literalFromConstantValue(value.Value); literal != nil {
				return literal
			}
		}
	}
	if ident, ok := expr.(*ast.Ident); ok {
		if object := g.getObject(ident); object != nil {
			if constant, ok := object.(*types.Const); ok {
				return literalFromConstantValue(constant.Val())
			}
		}
	}
	return nil
}

func packageOwnsObject(pkg *packages.Package, object types.Object) bool {
	if pkg == nil || object == nil {
		return false
	}
	owner := object.Pkg()
	return owner != nil && pkg.PkgPath == owner.Path()
}

func (g *GoFile) declarationForObject(object types.Object) (ast.Expr, token.Pos, bool) {
	if g == nil || object == nil {
		return nil, token.NoPos, false
	}
	for _, pkg := range g.ImportedPackages {
		if !packageOwnsObject(pkg, object) || pkg.TypesInfo == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			if file == nil {
				continue
			}
			var declaration ast.Expr
			var declarationPos token.Pos
			ast.Inspect(file, func(node ast.Node) bool {
				if declaration != nil {
					return false
				}
				switch node := node.(type) {
				case *ast.ValueSpec:
					for idx, name := range node.Names {
						if pkg.TypesInfo.Defs[name] == object && idx < len(node.Values) {
							declaration = node.Values[idx]
							declarationPos = node.Pos()
							return false
						}
					}
				case *ast.AssignStmt:
					for idx, lhs := range node.Lhs {
						if id, ok := lhs.(*ast.Ident); ok && pkg.TypesInfo.Defs[id] == object && idx < len(node.Rhs) {
							declaration = node.Rhs[idx]
							declarationPos = node.Pos()
							return false
						}
					}
				}
				return true
			})
			if declaration != nil {
				return declaration, declarationPos, true
			}
		}
	}
	return nil, token.NoPos, false
}

func (g *GoFile) getObject(ident *ast.Ident) types.Object {
	if g == nil || ident == nil {
		return nil
	}
	for _, pkg := range g.ImportedPackages {
		if pkg == nil || pkg.TypesInfo == nil {
			continue
		}
		if object := pkg.TypesInfo.Uses[ident]; object != nil {
			return object
		}
		if object := pkg.TypesInfo.Defs[ident]; object != nil {
			return object
		}
	}
	return nil
}

// isMutatedBefore reports whether a variable has been assigned a new value before its use.
func (g *GoFile) isMutatedBefore(object types.Object, before token.Pos) bool {
	if g == nil {
		return false
	}
	if _, ok := object.(*types.Var); !ok {
		return false
	}

	for _, pkg := range g.ImportedPackages {
		if !packageOwnsObject(pkg, object) || pkg.TypesInfo == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			if file == nil {
				continue
			}
			for node := range ast.Preorder(file) {
				switch node := node.(type) {
				case *ast.AssignStmt:
					if node.Pos() >= before {
						continue
					}
					for _, lhs := range node.Lhs {
						if ident, ok := lhs.(*ast.Ident); ok && pkg.TypesInfo.Uses[ident] == object {
							return true
						}
					}

				case *ast.IncDecStmt:
					if node.Pos() >= before {
						continue
					}
					if ident, ok := node.X.(*ast.Ident); ok && pkg.TypesInfo.Uses[ident] == object {
						return true
					}

				case *ast.RangeStmt:
					if node.Body == nil || node.Body.Pos() >= before {
						continue
					}
					for _, candidate := range []ast.Expr{node.Key, node.Value} {
						ident, ok := candidate.(*ast.Ident)
						if !ok {
							continue
						}
						switch node.Tok {
						case token.DEFINE:
							if pkg.TypesInfo.Defs[ident] == object {
								return true
							}
						case token.ASSIGN:
							if pkg.TypesInfo.Uses[ident] == object {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}
