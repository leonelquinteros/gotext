package dir

import (
	"go/ast"
	"go/token"
	"log"
	"strconv"

	"golang.org/x/tools/go/packages"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

// register go parser
func init() {
	AddParser(goParser)
}

// parse go package
func goParser(dirPath, basePath string, data *parser.DomainMap) error {
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
			parser.GoFile{
				PkgConf:  &conf,
				FilePath: fileSet.Position(node.Package).Filename,
				BasePath: basePath,
				Data:     data,
				FileSet:  fileSet,

				ImportedPackages: map[string]*packages.Package{
					pkgs[0].Name: pkgs[0],
				},
			},
		}

		ast.Inspect(node, file.InspectFile)
	}
	return nil
}

// GoFile handles the parsing of one go file
type GoFile struct {
	parser.GoFile
}

// getPackage loads module by name
func (g *GoFile) GetPackage(name string) (*packages.Package, error) {
	pkgs, err := packages.Load(g.PkgConf, name)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil
	}
	return pkgs[0], nil
}

func (g *GoFile) InspectFile(n ast.Node) bool {
	switch x := n.(type) {
	// get names of imported packages
	case *ast.ImportSpec:
		packageName, _ := strconv.Unquote(x.Path.Value)

		pkg, err := g.GetPackage(packageName)
		if err != nil {
			log.Printf("failed to load package %s: %s", packageName, err)
		} else {
			if x.Name == nil {
				g.ImportedPackages[pkg.Name] = pkg
			} else {
				g.ImportedPackages[x.Name.Name] = pkg
			}
		}

	// check each function call
	case *ast.CallExpr:
		g.InspectCallExpr(x)

	default:
		print()
	}

	return true
}
