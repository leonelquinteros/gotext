package pkgtree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
)

func TestParsePkgTree(t *testing.T) {
	defaultDomain := "default"
	data := &parser.DomainMap{
		Default: defaultDomain,
	}
	currentPath, err := os.Getwd()
	pkgPath := filepath.Join(filepath.Dir(filepath.Dir(currentPath)), "fixtures")
	println(pkgPath)
	if err != nil {
		t.Error(err)
	}
	err = ParsePkgTree(pkgPath, data, true)
	if err != nil {
		t.Error(err)
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
