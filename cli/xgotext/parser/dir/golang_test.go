package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

func TestGoParserExtractsStaticStrings(t *testing.T) {
	data := &parser.DomainMap{Default: "default"}
	fixtures := filepath.Join("..", "..", "fixtures")

	if err := goParser(fixtures, fixtures, data); err != nil {
		t.Fatal(err)
	}

	if _, ok := data.Domains["constants"].Translations["message from a constant"]; !ok {
		t.Error("constant domain and message were not extracted")
	}
	if plural := data.Domains["default"].Translations["singular from a constant"]; plural == nil || plural.MsgIDPlural != "plural from a constant" {
		t.Error("constant plural strings were not extracted")
	}
	if _, ok := data.Domains["default"].ContextTranslations["constant context"]["message with a constant context"]; !ok {
		t.Error("constant context and message were not extracted")
	}
	if _, ok := data.Domains["default"].Translations["message before mutation"]; ok {
		t.Error("mutated variable should not be extracted")
	}
	if _, ok := data.Domains["default"].Translations["message after mutation"]; ok {
		t.Error("mutated variable should not be extracted")
	}
	if _, ok := data.Domains["default"].Translations["message with a dynamic domain"]; ok {
		t.Error("dynamic domain should not be extracted")
	}
}

func TestGoParserReturnsDiagnosticsBeforeExtraction(t *testing.T) {
	root := writeDirectoryModule(t)
	if err := os.WriteFile(filepath.Join(root, "good.go"), []byte(`package fixture

import "github.com/leonelquinteros/gotext"

func good() { gotext.Get("must not be extracted") }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte("package fixture\n\nfunc broken(\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	data := &parser.DomainMap{Default: "default"}
	if err := goParser(root, root, data); err == nil {
		t.Fatal("goParser unexpectedly succeeded for malformed Go source")
	}
	if len(data.Domains) != 0 {
		t.Fatalf("goParser mutated DomainMap after diagnostics: %#v", data.Domains)
	}
}

func TestParseDirRecSkipsEmptyAndAssetsOnlyDirectories(t *testing.T) {
	root := writeDirectoryModule(t)
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(`package fixture

import "github.com/leonelquinteros/gotext"

func healthy() { gotext.Get("healthy directory") }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"empty", "assets"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "readme.txt"), []byte("asset"), 0o644); err != nil {
		t.Fatal(err)
	}

	data := &parser.DomainMap{Default: "default"}
	if err := ParseDirRec(root, nil, data, false); err != nil {
		t.Fatalf("ParseDirRec failed for healthy module with non-Go directories: %v", err)
	}
	domain := data.Domains["default"]
	if domain == nil || domain.Translations["healthy directory"] == nil {
		t.Fatal("healthy Go source was not extracted")
	}
}

func TestGoParserRespectsBuildExcludedSources(t *testing.T) {
	root := writeDirectoryModule(t)
	if err := os.WriteFile(filepath.Join(root, "tagged.go"), []byte(`//go:build gotextaudit

package fixture

import "github.com/leonelquinteros/gotext"

func tagged() { gotext.Get("tagged message") }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFLAGS", "")
	data := &parser.DomainMap{Default: "default"}
	if err := goParser(root, root, data); err != nil {
		t.Fatalf("build-excluded directory should be skipped: %v", err)
	}
	if len(data.Domains) != 0 {
		t.Fatalf("excluded source produced translations: %#v", data.Domains)
	}
	t.Setenv("GOFLAGS", "-tags=gotextaudit")
	if err := goParser(root, root, data); err != nil {
		t.Fatalf("enabled source failed: %v", err)
	}
	if domain := data.Domains["default"]; domain == nil || domain.Translations["tagged message"] == nil {
		t.Fatal("enabled source did not produce its translation")
	}
}

func writeDirectoryModule(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	module := fmt.Sprintf(`module example.com/dirfixture

go 1.27

require github.com/leonelquinteros/gotext v0.0.0

replace github.com/leonelquinteros/gotext => %s
`, repositoryRoot(t))
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(module), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}
