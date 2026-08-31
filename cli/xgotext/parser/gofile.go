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
	for _, pkg := range g.ImportedPackages {
		if pkg.TypesInfo == nil {
			continue
		}
		if obj, ok := pkg.TypesInfo.Uses[ident]; ok {
			return obj
		}
	}
	return nil
}

// getExprType from any expression
func (g *GoFile) getExprType(expr ast.Expr) types.Type {
	for _, pkg := range g.ImportedPackages {
		if pkg.TypesInfo == nil {
			continue
		}
		if tv, ok := pkg.TypesInfo.Types[expr]; ok {
			return tv.Type
		}
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
		if t.Obj().Pkg() == nil || t.Obj().Pkg().Path() != "github.com/leonelquinteros/gotext" {
			return false
		}

	case *types.Alias:
		return g.CheckType(t.Rhs())

	case *types.Interface:
		// Check if it's the Translator interface from our package
		// This is used for interfaces like 'Translator'
		return t.NumMethods() > 0 && t.Method(0).Pkg() != nil && t.Method(0).Pkg().Path() == "github.com/leonelquinteros/gotext"

	default:
		return false
	}
	return true
}

// InspectCallExpr inspects the call expression
func (g *GoFile) InspectCallExpr(n *ast.CallExpr) {
	// must be a selector expression otherwise it is a local function call
	expr, ok := n.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// Resolve the type of the receiver (expr.X)
	receiverType := g.getExprType(expr.X)
	if receiverType == nil {
		// Fallback for package calls if types didn't resolve it
		if id, ok := expr.X.(*ast.Ident); ok && id.Obj == nil {
			pkg, ok := g.ImportedPackages[id.Name]
			if !ok || pkg.PkgPath != "github.com/leonelquinteros/gotext" {
				return
			}
		} else {
			return
		}
	} else if !g.CheckType(receiverType) {
		return
	}

	// convert args
	args := make([]*ast.BasicLit, len(n.Args))
	resolving := make(map[types.Object]bool)
	for idx, arg := range n.Args {
		args[idx] = g.resolveStringLiteral(arg, n.Pos(), resolving)
	}

	// get position
	path, _ := filepath.Rel(g.BasePath, g.FilePath)
	position := fmt.Sprintf("%s:%d", path, g.FileSet.Position(n.Lparen).Line)

	// handle getters
	if def, ok := gotextGetter[expr.Sel.String()]; ok {
		g.ParseGetter(def, args, position)
		return
	}
}

// ParseGetter parses the getter function
func (g *GoFile) ParseGetter(def GetterDef, args []*ast.BasicLit, pos string) {
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
	if def.Plural > 0 {
		// plural ID must be a string
		msgIDPlural, ok := getStringArgument(args, def.Plural, "plural", pos)
		if !ok {
			return
		}
		trans.MsgIDPlural = msgIDPlural
	}
	if def.Context > 0 {
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
	if literal, ok := expr.(*ast.BasicLit); ok && literal.Kind == token.STRING {
		return literal
	}

	if literal := g.constantStringLiteral(expr); literal != nil {
		return literal
	}

	ident, ok := expr.(*ast.Ident)
	if !ok || ident.Obj == nil {
		return nil
	}
	object := g.getObject(ident)
	if object != nil {
		if resolving[object] || g.isMutatedBefore(object, before) {
			return nil
		}
		resolving[object] = true
		defer delete(resolving, object)
	}

	switch decl := ident.Obj.Decl.(type) {
	case *ast.ValueSpec:
		for i, name := range decl.Names {
			if name.Name == ident.Name && i < len(decl.Values) {
				return g.resolveStringLiteral(decl.Values[i], decl.Pos(), resolving)
			}
		}
	case *ast.AssignStmt:
		for i, lhs := range decl.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name == ident.Name && i < len(decl.Rhs) {
				return g.resolveStringLiteral(decl.Rhs[i], decl.Pos(), resolving)
			}
		}
	}
	return nil
}

// getLiteralFromIdent resolves a declaration without package type information.
// It is retained for focused AST tests; production extraction uses resolveStringLiteral.
func getLiteralFromIdent(ident *ast.Ident) *ast.BasicLit {
	return (&GoFile{}).resolveStringLiteral(ident, token.NoPos, make(map[types.Object]bool))
}

func (g *GoFile) constantStringLiteral(expr ast.Expr) *ast.BasicLit {
	for _, pkg := range g.ImportedPackages {
		if pkg.TypesInfo == nil {
			continue
		}
		if value, ok := pkg.TypesInfo.Types[expr]; ok && value.Value != nil && value.Value.Kind() == constant.String {
			return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(constant.StringVal(value.Value))}
		}
	}
	return nil
}

func (g *GoFile) getObject(ident *ast.Ident) types.Object {
	for _, pkg := range g.ImportedPackages {
		if pkg.TypesInfo == nil {
			continue
		}
		if object := pkg.TypesInfo.Uses[ident]; object != nil {
			return object
		}
	}
	return nil
}

// isMutatedBefore reports whether a variable has been assigned a new value before its use.
func (g *GoFile) isMutatedBefore(object types.Object, before token.Pos) bool {
	if _, ok := object.(*types.Var); !ok {
		return false
	}

	for _, pkg := range g.ImportedPackages {
		if pkg.TypesInfo == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			for node := range ast.Preorder(file) {
				assign, ok := node.(*ast.AssignStmt)
				if !ok || assign.Pos() >= before {
					continue
				}
				for _, lhs := range assign.Lhs {
					ident, ok := lhs.(*ast.Ident)
					if ok && pkg.TypesInfo.Uses[ident] == object {
						return true
					}
				}
			}
		}
	}
	return false
}
