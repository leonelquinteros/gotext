package pkg_tree

import (
	"github.com/leonelquinteros/gotext/cli/xgotext/parser"
	"os"
	"path/filepath"
	"testing"
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

	translations := []string{"\"inside sub package\"", "\"My text on 'domain-name' domain\"", "\"alias call\"", "\"Singular\"", "\"SingularVar\"", "\"translate package\"", "\"translate sub package\"", "\"inside dummy\""}

	if len(translations) != len(data.Domains[defaultDomain].Translations) {
		t.Error("translations count mismatch")
	}
	for _, tr := range translations {
		if _, ok := data.Domains[defaultDomain].Translations[tr]; !ok {
			t.Errorf("translation '%v' not in result", tr)
		}
	}
}
