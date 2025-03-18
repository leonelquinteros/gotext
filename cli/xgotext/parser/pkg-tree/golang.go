package pkg_tree

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"strconv"

	"golang.org/x/tools/go/packages"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

// ParsePkgTree parse go package tree
func ParsePkgTree(pkgPath string, data *parser.DomainMap, verbose bool) error {
	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return pkgParser(pkgPath, basePath, data, verbose)
}

func pkgParser(dirPath, basePath string, data *parser.DomainMap, verbose bool) error {
	mainPkg, err := loadPackage(dirPath)
	if err != nil {
		return err
	}

	for _, pkg := range filterPkgs(mainPkg) {
		if verbose {
			fmt.Println(pkg.ID)
		}
		for _, node := range pkg.Syntax {
			file := GoFile{
				parser.GoFile{
					FilePath: pkg.Fset.Position(node.Package).Filename,
					BasePath: basePath,
					Data:     data,
					FileSet:  pkg.Fset,

					ImportedPackages: map[string]*packages.Package{
						pkg.Name: pkg,
					},
				},
			}

			ast.Inspect(node, file.InspectFile)
		}
	}

	return nil
}

var pkgCache = make(map[string]*packages.Package)

func loadPackage(name string) (*packages.Package, error) {
	fileSet := token.NewFileSet()
	conf := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedDeps,
		Fset: fileSet,
		Dir:  name,
	}
	pkgs, err := packages.Load(conf)
	if err != nil {
		return nil, err
	}

	return pkgs[0], nil
}

func filterPkgs(pkg *packages.Package) []*packages.Package {
	result := filterPkgsRec(pkg)
	return result
}

func filterPkgsRec(pkg *packages.Package) []*packages.Package {
	result := make([]*packages.Package, 0, 100)
	pkgCache[pkg.ID] = pkg
	for _, importedPkg := range pkg.Imports {
		if importedPkg.ID == "github.com/leonelquinteros/gotext" {
			result = append(result, pkg)
		}
		if _, ok := pkgCache[importedPkg.ID]; ok {
			continue
		}
		result = append(result, filterPkgsRec(importedPkg)...)
	}
	return result
}

// GoFile handles the parsing of one go file
type GoFile struct {
	parser.GoFile
}

// GetPackage loads module by name
func (g *GoFile) GetPackage(name string) (*packages.Package, error) {
	pkg, ok := pkgCache[name]
	if !ok {
		return nil, fmt.Errorf("not found in cache")
	}
	return pkg, nil
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
