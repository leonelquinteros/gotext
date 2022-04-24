package pkg_tree

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

	translations := []string{
		"\"inside sub package\"",
		"\"My text on 'domain-name' domain\"",
		"\"This is a string addition. Which is merged.\"",
		"\"\"\n\"This is a string\\n\"\n\"with an Newline\"",
		"\"\"\n\"This is a multiline string.\\n\"\n\"It should be formatted properly in a .pot file.\"",
		"\"alias call\"",
		"\"Singular\"",
		"\"SingularVar\"",
		"\"translate package\"",
		"\"translate sub package\"",
		"\"inside dummy\"",
	}

	if len(translations) != len(data.Domains[defaultDomain].Translations) {
		t.Error("translations count mismatch")
	}
	for _, tr := range translations {
		if _, ok := data.Domains[defaultDomain].Translations[tr]; !ok {
			t.Errorf("translation '%v' not in result", tr)
		}
	}
}
