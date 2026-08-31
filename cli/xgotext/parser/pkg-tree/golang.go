package pkgtree

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"strconv"
	"sync"

	"golang.org/x/tools/go/packages"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

// ParsePkgTree parse go package tree
func ParsePkgTree(pkgPath string, data *parser.DomainMap, verbose bool) error {
	basePath, err := os.Getwd()
	if err != nil {
		return err
	}
	resetPackageCache()
	return pkgParser(pkgPath, basePath, data, verbose)
}

func pkgParser(dirPath, basePath string, data *parser.DomainMap, verbose bool) error {
	mainPkg, err := loadPackage(dirPath)
	if err != nil {
		return err
	}

	pkgs, cache := filterPkgsScoped(mainPkg)
	for _, pkg := range pkgs {
		if verbose {
			fmt.Println(pkg.ID)
		}
		for _, node := range pkg.Syntax {
			file := GoFile{
				GoFile: parser.GoFile{
					FilePath: pkg.Fset.Position(node.Package).Filename,
					BasePath: basePath,
					Data:     data,
					FileSet:  pkg.Fset,

					ImportedPackages: map[string]*packages.Package{
						pkg.Name: pkg,
					},
				},
				packageCache: cache,
			}

			ast.Inspect(node, file.InspectFile)
		}
	}

	return nil
}

var (
	pkgCache   = make(map[string]*packages.Package)
	pkgCacheMu sync.RWMutex
)

func resetPackageCache() {
	pkgCacheMu.Lock()
	pkgCache = make(map[string]*packages.Package)
	pkgCacheMu.Unlock()
}

func setPackageCache(cache map[string]*packages.Package) {
	pkgCacheMu.Lock()
	pkgCache = cache
	pkgCacheMu.Unlock()
}

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
	if err := packageDiagnostics(pkgs); err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for %q", name)
	}

	return pkgs[0], nil
}

func packageDiagnostics(pkgs []*packages.Package) error {
	seen := make(map[*packages.Package]bool, len(pkgs))
	var visit func(*packages.Package) error
	visit = func(pkg *packages.Package) error {
		if pkg == nil {
			return fmt.Errorf("package loader returned a nil package")
		}
		if seen[pkg] {
			return nil
		}
		seen[pkg] = true
		if len(pkg.Errors) > 0 {
			return fmt.Errorf("package %q has diagnostics: %v", pkg.ID, pkg.Errors)
		}
		for _, importedPkg := range pkg.Imports {
			if err := visit(importedPkg); err != nil {
				return err
			}
		}
		return nil
	}

	for _, pkg := range pkgs {
		if err := visit(pkg); err != nil {
			return err
		}
	}
	return nil
}

func filterPkgs(pkg *packages.Package) []*packages.Package {
	result, cache := filterPkgsScoped(pkg)
	setPackageCache(cache)
	return result
}

func filterPkgsScoped(pkg *packages.Package) ([]*packages.Package, map[string]*packages.Package) {
	cache := make(map[string]*packages.Package, 100)
	return filterPkgsRecWithCache(pkg, cache), cache
}

func filterPkgsRecWithCache(pkg *packages.Package, cache map[string]*packages.Package) []*packages.Package {
	if pkg == nil {
		return nil
	}
	if _, ok := cache[pkg.ID]; ok {
		return nil
	}
	cache[pkg.ID] = pkg

	result := make([]*packages.Package, 0, 100)
	for _, importedPkg := range pkg.Imports {
		if importedPkg == nil {
			continue
		}
		if importedPkg.ID == "github.com/leonelquinteros/gotext" {
			result = append(result, pkg)
		}
		if _, ok := cache[importedPkg.ID]; ok {
			continue
		}
		result = append(result, filterPkgsRecWithCache(importedPkg, cache)...)
	}
	return result
}

// GoFile handles the parsing of one go file
type GoFile struct {
	parser.GoFile
	packageCache map[string]*packages.Package
}

// GetPackage loads module by name
func (g *GoFile) GetPackage(name string) (*packages.Package, error) {
	if g.packageCache != nil {
		if pkg, ok := g.packageCache[name]; ok {
			return pkg, nil
		}
		return nil, fmt.Errorf("not found in cache")
	}

	pkgCacheMu.RLock()
	pkg, ok := pkgCache[name]
	pkgCacheMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("not found in cache")
	}
	return pkg, nil
}

// InspectFile inspects the AST node
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
