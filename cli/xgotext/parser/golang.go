package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"strconv"

	"golang.org/x/tools/go/packages"
)

// GetterDef describes a getter
type GetterDef struct {
	Id      int
	Plural  int
	Context int
	Domain  int
}

// maxArgIndex returns the largest argument index
func (d *GetterDef) maxArgIndex() int {
	m := d.Id
	if d.Plural > m {
		m = d.Plural
	}
	if d.Context > m {
		m = d.Context
	}
	if d.Domain > m {
		m = d.Domain
	}
	return m
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

// register go parser
func init() {
	AddParser(goParser)
}

// parse go package
func goParser(dirPath, basePath string, data *DomainMap) error {
	fileSet := token.NewFileSet()

	conf := packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Fset: fileSet,
		Dir:  basePath,
	}

	// load package from path
	pkgs, err := packages.Load(&packages.Config{
		Mode: conf.Mode,
		Fset: fileSet,
		Dir:  dirPath,
	})
	if err != nil || len(pkgs) == 0 {
		// not a go package
		return nil
	}

	// handle each file
	for _, node := range pkgs[0].Syntax {
		file := GoFile{
			pkgConf:  &conf,
			filePath: fileSet.Position(node.Package).Filename,
			basePath: basePath,
			data:     data,
			fileSet:  fileSet,

			importedPackages: map[string]*packages.Package{
				pkgs[0].Name: pkgs[0],
			},
		}

		ast.Inspect(node, file.inspectFile)
	}
	return nil
}

// GoFile handles the parsing of one go file
type GoFile struct {
	filePath string
	basePath string
	data     *DomainMap

	fileSet *token.FileSet
	pkgConf *packages.Config

	importedPackages map[string]*packages.Package
}

// getPackage loads module by name
func (g *GoFile) getPackage(name string) (*packages.Package, error) {
	pkgs, err := packages.Load(g.pkgConf, name)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil
	}
	return pkgs[0], nil
}

// getType from ident object
func (g *GoFile) getType(ident *ast.Ident) types.Object {
	for _, pkg := range g.importedPackages {
		if obj, ok := pkg.TypesInfo.Uses[ident]; ok {
			return obj
		}
	}
	return nil
}

func (g *GoFile) inspectFile(n ast.Node) bool {
	switch x := n.(type) {
	// get names of imported packages
	case *ast.ImportSpec:
		packageName, _ := strconv.Unquote(x.Path.Value)

		pkg, err := g.getPackage(packageName)
		if err != nil {
			log.Printf("failed to load package %s: %s", packageName, err)
		} else {
			if x.Name == nil {
				g.importedPackages[pkg.Name] = pkg
			} else {
				g.importedPackages[x.Name.Name] = pkg
			}
		}

	// check each function call
	case *ast.CallExpr:
		g.inspectCallExpr(x)

	default:
		print()
	}

	return true
}

// checkType for gotext object
func (g *GoFile) checkType(rawType types.Type) bool {
	switch t := rawType.(type) {
	case *types.Pointer:
		return g.checkType(t.Elem())

	case *types.Named:
		if t.Obj().Pkg() == nil || t.Obj().Pkg().Path() != "github.com/leonelquinteros/gotext" {
			return false
		}
	default:
		return false
	}
	return true
}

func (g *GoFile) inspectCallExpr(n *ast.CallExpr) {
	// must be a selector expression otherwise it is a local function call
	expr, ok := n.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	switch e := expr.X.(type) {
	// direct call
	case *ast.Ident:
		// object is a package if the Obj is not set
		if e.Obj == nil {
			pkg, ok := g.importedPackages[e.Name]
			if !ok || pkg.PkgPath != "github.com/leonelquinteros/gotext" {
				return
			}

		} else {
			// validate type of object
			t := g.getType(e)
			if t == nil || !g.checkType(t.Type()) {
				return
			}
		}

	// call to attribute
	case *ast.SelectorExpr:
		// validate type of object
		t := g.getType(e.Sel)
		if t == nil || !g.checkType(t.Type()) {
			return
		}

	default:
		return
	}

	// convert args
	args := make([]*ast.BasicLit, len(n.Args))
	for idx, arg := range n.Args {
		args[idx], _ = arg.(*ast.BasicLit)
	}

	// get position
	path, _ := filepath.Rel(g.basePath, g.filePath)
	position := fmt.Sprintf("%s:%d", path, g.fileSet.Position(n.Lparen).Line)

	// handle getters
	if def, ok := gotextGetter[expr.Sel.String()]; ok {
		g.parseGetter(def, args, position)
		return
	}
}

func (g *GoFile) parseGetter(def GetterDef, args []*ast.BasicLit, pos string) {
	// check if enough arguments are given
	if len(args) < def.maxArgIndex() {
		return
	}

	// get domain
	var domain string
	if def.Domain != -1 {
		domain, _ = strconv.Unquote(args[def.Domain].Value)
	}

	// only handle function calls with strings as ID
	if args[def.Id] == nil || args[def.Id].Kind != token.STRING {
		log.Printf("ERR: Unsupported call at %s (ID not a string)", pos)
		return
	}

	trans := Translation{
		MsgId:           args[def.Id].Value,
		SourceLocations: []string{pos},
	}
	if def.Plural > 0 {
		// plural ID must be a string
		if args[def.Plural] == nil || args[def.Plural].Kind != token.STRING {
			log.Printf("ERR: Unsupported call at %s (Plural not a string)", pos)
			return
		}
		trans.MsgIdPlural = args[def.Plural].Value
	}
	if def.Context > 0 {
		// Context must be a string
		if args[def.Context] == nil || args[def.Context].Kind != token.STRING {
			log.Printf("ERR: Unsupported call at %s (Context not a string)", pos)
			return
		}
		trans.Context = args[def.Context].Value
	}

	g.data.AddTranslation(domain, &trans)
}
