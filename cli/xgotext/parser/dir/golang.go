package dir

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"path/filepath"
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
	hasSource, err := hasGoSource(dirPath)
	if err != nil {
		return fmt.Errorf("failed to inspect Go source in %q: %w", dirPath, err)
	}
	if !hasSource {
		return nil
	}

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
	loadConf := conf
	loadConf.Dir = dirPath
	pkgs, err := packages.Load(&loadConf)
	if err != nil {
		return fmt.Errorf("failed to load Go package in %q: %w", dirPath, err)
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("failed to load Go package in %q: no packages found", dirPath)
	}
	if len(pkgs) == 1 && pkgs[0] != nil && len(pkgs[0].GoFiles) == 0 {
		for _, filename := range pkgs[0].IgnoredFiles {
			if filepath.Ext(filename) == ".go" {
				return nil // The active build excludes this directory's Go source.
			}
		}
	}
	if err := packageDiagnostics(pkgs); err != nil {
		return fmt.Errorf("failed to load Go package in %q: %w", dirPath, err)
	}

	pkg := pkgs[0]

	// handle each file
	for _, node := range pkg.Syntax {
		if node == nil {
			continue
		}
		file := GoFile{
			parser.GoFile{
				PkgConf:  &conf,
				FilePath: fileSet.Position(node.Package).Filename,
				BasePath: basePath,
				Data:     data,
				FileSet:  fileSet,

				ImportedPackages: map[string]*packages.Package{
					pkg.Name: pkg,
				},
			},
		}

		ast.Inspect(node, file.InspectFile)
	}
	return nil
}

func hasGoSource(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
			return true, nil
		}
	}
	return false, nil
}

func packageDiagnostics(pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		if pkg == nil {
			return fmt.Errorf("package loader returned a nil package")
		}
		if len(pkg.Errors) > 0 {
			return fmt.Errorf("package %q has diagnostics: %v", pkg.ID, pkg.Errors)
		}
	}
	return nil
}

// GoFile handles the parsing of one go file
type GoFile struct {
	parser.GoFile
}

// GetPackage loads module by name
func (g *GoFile) GetPackage(name string) (*packages.Package, error) {
	if g == nil {
		return nil, fmt.Errorf("failed to load package %q from a nil GoFile", name)
	}
	pkgs, err := packages.Load(g.PkgConf, name)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %q: %w", name, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for %q", name)
	}
	if err := packageDiagnostics(pkgs); err != nil {
		return nil, fmt.Errorf("failed to load package %q: %w", name, err)
	}
	return pkgs[0], nil
}

// InspectFile inspects the file node
func (g *GoFile) InspectFile(n ast.Node) bool {
	switch x := n.(type) {
	// get names of imported packages
	case *ast.ImportSpec:
		packageName, err := strconv.Unquote(x.Path.Value)
		if err != nil {
			log.Printf("failed to decode package import %q: %s", x.Path.Value, err)
			break
		}

		pkg, err := g.GetPackage(packageName)
		if err != nil {
			log.Printf("failed to load package %s: %s", packageName, err)
		} else if pkg == nil {
			log.Printf("failed to load package %s: package loader returned a nil package", packageName)
		} else {
			if g.ImportedPackages == nil {
				g.ImportedPackages = make(map[string]*packages.Package)
			}
			if x.Name == nil {
				g.ImportedPackages[pkg.Name] = pkg
			} else {
				g.ImportedPackages[x.Name.Name] = pkg
			}
		}

	// check each function call
	case *ast.CallExpr:
		g.InspectCallExpr(x)

	}

	return true
}
