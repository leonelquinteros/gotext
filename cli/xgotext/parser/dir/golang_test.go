package dir

import (
	"path/filepath"
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
	if _, ok := data.Domains["default"].ContextTranslations[`"constant context"`]["message with a constant context"]; !ok {
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
