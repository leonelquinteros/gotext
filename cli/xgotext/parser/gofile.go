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
func (d *GetterDef) MaxArgIndex() int {
	return max(d.Id, d.Plural, d.Context, d.Domain)
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
		if pkg.Types == nil {
			continue
		}
		if obj, ok := pkg.TypesInfo.Uses[ident]; ok {
			return obj
		}
	}
	return nil
}

// checkType for gotext object
func (g *GoFile) CheckType(rawType types.Type) bool {
	switch t := rawType.(type) {
	case *types.Pointer:
		return g.CheckType(t.Elem())

	case *types.Named:
		if t.Obj().Pkg() == nil || t.Obj().Pkg().Path() != "github.com/leonelquinteros/gotext" {
			return false
		}

	case *types.Alias:
		return g.CheckType(t.Rhs())

	default:
		return false
	}
	return true
}

func (g *GoFile) InspectCallExpr(n *ast.CallExpr) {
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
			pkg, ok := g.ImportedPackages[e.Name]
			if !ok || pkg.PkgPath != "github.com/leonelquinteros/gotext" {
				return
			}

		} else {
			// validate type of object
			t := g.GetType(e)
			if t == nil || !g.CheckType(t.Type()) {
				return
			}
		}

	// call to attribute
	case *ast.SelectorExpr:
		// validate type of object
		t := g.GetType(e.Sel)
		if t == nil || !g.CheckType(t.Type()) {
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
	path, _ := filepath.Rel(g.BasePath, g.FilePath)
	position := fmt.Sprintf("%s:%d", path, g.FileSet.Position(n.Lparen).Line)

	// handle getters
	if def, ok := gotextGetter[expr.Sel.String()]; ok {
		g.ParseGetter(def, args, position)
		return
	}
}

func (g *GoFile) ParseGetter(def GetterDef, args []*ast.BasicLit, pos string) {
	// check if enough arguments are given
	if len(args) < def.MaxArgIndex() {
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

	msgID, _ := strconv.Unquote(args[def.Id].Value)
	trans := Translation{
		MsgId:           msgID,
		SourceLocations: []string{pos},
	}
	if def.Plural > 0 {
		// plural ID must be a string
		if args[def.Plural] == nil || args[def.Plural].Kind != token.STRING {
			log.Printf("ERR: Unsupported call at %s (Plural not a string)", pos)
			return
		}
		msgIDPlural, _ := strconv.Unquote(args[def.Plural].Value)
		trans.MsgIdPlural = msgIDPlural
	}
	if def.Context > 0 {
		// Context must be a string
		if args[def.Context] == nil || args[def.Context].Kind != token.STRING {
			log.Printf("ERR: Unsupported call at %s (Context not a string)", pos)
			return
		}
		trans.Context = args[def.Context].Value
	}

	g.Data.AddTranslation(domain, &trans)
}
