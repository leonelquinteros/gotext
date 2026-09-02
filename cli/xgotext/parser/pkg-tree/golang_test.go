package pkgtree

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"

	"golang.org/x/tools/go/packages"
)

func TestParsePkgTree(t *testing.T) {
	defaultDomain := "default"
	data := &parser.DomainMap{
		Default: defaultDomain,
	}
	pkgPath := filepath.Join(repositoryRoot(t), "cli", "xgotext", "fixtures")
	if err := ParsePkgTree(pkgPath, data, true); err != nil {
		t.Fatal(err)
	}

	translations := []string{"inside sub package", "My text on 'domain-name' domain", "alias call", "Singular", "SingularVar", "translate package", "translate sub package", "inside dummy",
		`string with backquotes`, "string ending with EOL\n", "string with\nmultiple\nEOL", `raw string with\nmultiple\nEOL`,
		`multi
line
string`,
		`multi
line
string
ending with
EOL`, "multline\nending with EOL\n", "type alias", "locale constructor call",
		"chained locale", "chained po", "chained from func", "from interface",
		"this is constant testing",
		"this is multi const 1",
		"this is multi const 2",
		"this is variable testing",
		"this is multi var 1",
		"this is multi var 2",
		"some string to translate",
		"message from a constant",
		"concatenated constant",
		"singular from a constant",
	}

	if len(translations) != len(data.Domains[defaultDomain].Translations) {
		t.Error("translations count mismatch")
	}
	for _, tr := range translations {
		if _, ok := data.Domains[defaultDomain].Translations[tr]; !ok {
			t.Errorf("translation '%v' not in result", tr)
		}
	}
	if _, ok := data.Domains["constants"].Translations["message from a constant"]; !ok {
		t.Error("constant domain and message were not extracted")
	}
	if plural := data.Domains[defaultDomain].Translations["singular from a constant"]; plural == nil || plural.MsgIDPlural != "plural from a constant" {
		t.Error("constant plural strings were not extracted")
	}
	if _, ok := data.Domains[defaultDomain].ContextTranslations[`"constant context"`]["message with a constant context"]; !ok {
		t.Error("constant context and message were not extracted")
	}
	if _, ok := data.Domains[defaultDomain].Translations["message before mutation"]; ok {
		t.Error("mutated variable should not be extracted")
	}
	if _, ok := data.Domains[defaultDomain].Translations["message after mutation"]; ok {
		t.Error("mutated variable should not be extracted")
	}
	if _, ok := data.Domains[defaultDomain].Translations["message with a dynamic domain"]; ok {
		t.Error("dynamic domain should not be extracted")
	}
}
func TestLoadPackageReturnsErrorForMissingRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if _, err := loadPackage(missing); err == nil {
		t.Fatal("loadPackage unexpectedly succeeded without a package")
	}
}

func TestLoadPackageReturnsPackageDiagnostics(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/broken\n\ngo 1.27\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "broken.go"), []byte("package broken\n\nfunc broken(\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadPackage(root)
	if err == nil {
		t.Fatal("loadPackage unexpectedly ignored package diagnostics")
	}
	if !strings.Contains(err.Error(), `package "example.com/broken" has diagnostics`) {
		t.Fatalf("loadPackage error = %v, want package diagnostics", err)
	}
}

func TestFilterPkgsCacheScope(t *testing.T) {
	const gotextID = "github.com/leonelquinteros/gotext"
	firstGotext := &packages.Package{ID: gotextID}
	first := &packages.Package{
		ID: "example.com/first",
		Imports: map[string]*packages.Package{
			gotextID: firstGotext,
		},
	}
	secondGotext := &packages.Package{ID: gotextID}
	second := &packages.Package{
		ID: "example.com/second",
		Imports: map[string]*packages.Package{
			gotextID: secondGotext,
		},
	}

	if got := filterPkgs(first); len(got) != 1 || got[0] != first {
		t.Fatalf("filterPkgs(first) = %v, want first package", got)
	}
	if got := filterPkgs(second); len(got) != 1 || got[0] != second {
		t.Fatalf("filterPkgs(second) = %v, want second package", got)
	}
	g := &GoFile{}
	if got, err := g.GetPackage(gotextID); err != nil || got != secondGotext {
		t.Fatalf("cached gotext package = %v, %v; want second graph package", got, err)
	}
}

func TestParsePkgTreeIndependentGraphs(t *testing.T) {
	repoRoot := repositoryRoot(t)
	firstRoot := writeTestGraph(t, repoRoot, "example.com/first", "first graph")
	secondRoot := writeTestGraph(t, repoRoot, "example.com/second", "second graph")

	firstData := &parser.DomainMap{Default: "default"}
	if err := ParsePkgTree(firstRoot, firstData, false); err != nil {
		t.Fatalf("first ParsePkgTree: %v", err)
	}
	if !hasTestTranslation(firstData, "first graph") {
		t.Fatal("first graph translation was not extracted")
	}
	if hasTestTranslation(firstData, "second graph") {
		t.Fatal("first graph reused a package from the second graph")
	}

	secondData := &parser.DomainMap{Default: "default"}
	if err := ParsePkgTree(secondRoot, secondData, false); err != nil {
		t.Fatalf("second ParsePkgTree: %v", err)
	}
	if !hasTestTranslation(secondData, "second graph") {
		t.Fatal("second graph translation was not extracted")
	}
	if hasTestTranslation(secondData, "first graph") {
		t.Fatal("second graph reused a package from the first graph")
	}
}
func writeTestGraph(t *testing.T, repoRoot, modulePath, message string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(fmt.Sprintf(`module %s

go 1.27

require github.com/leonelquinteros/gotext v0.0.0

replace github.com/leonelquinteros/gotext => %s
`, modulePath, repoRoot)), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "shared"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(fmt.Sprintf(`package app

import %q

func Run() { shared.Run() }
`, modulePath+"/shared")), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shared", "shared.go"), []byte(fmt.Sprintf(`package shared

import "github.com/leonelquinteros/gotext"

func Run() { gotext.Get(%q) }
`, message)), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func hasTestTranslation(data *parser.DomainMap, message string) bool {
	domain, ok := data.Domains["default"]
	if !ok || domain == nil {
		return false
	}
	_, ok = domain.Translations[message]
	return ok
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}
