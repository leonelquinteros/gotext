package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTranslation_AddLocations(t *testing.T) {
	tr := &Translation{
		MsgID: "test",
	}
	tr.AddLocations([]string{"file1.go:10"})
	if len(tr.SourceLocations) != 1 {
		t.Error("AddLocations failed to add to nil slice")
	}

	tr.AddLocations([]string{"file2.go:20"})
	if len(tr.SourceLocations) != 2 {
		t.Error("AddLocations failed to append")
	}
}

func TestTranslation_Dump(t *testing.T) {
	tr := &Translation{
		MsgID:           "test",
		SourceLocations: []string{"file.go:10"},
	}
	dump := tr.Dump()
	if !contains(dump, "msgid \"test\"") || !contains(dump, "#: file.go:10") {
		t.Error("Dump failed for simple translation")
	}

	tr.MsgIDPlural = "tests"
	dump = tr.Dump()
	if !contains(dump, "msgid_plural \"tests\"") || !contains(dump, "msgstr[0] \"\"") {
		t.Error("Dump failed for plural translation")
	}

	tr.Context = "ctx"
	dump = tr.Dump()
	if !contains(dump, "msgctxt ctx") {
		t.Error("Dump failed for context translation")
	}
}

func TestDomain_AddTranslation(t *testing.T) {
	d := &Domain{}
	tr := &Translation{
		MsgID:           "test",
		SourceLocations: []string{"file.go:10"},
	}
	d.AddTranslation(tr)
	if len(d.Translations) != 1 {
		t.Error("AddTranslation failed")
	}

	// Add same ID different location
	tr2 := &Translation{
		MsgID:           "test",
		SourceLocations: []string{"file.go:20"},
	}
	d.AddTranslation(tr2)
	if len(d.Translations) != 1 || len(d.Translations["test"].SourceLocations) != 2 {
		t.Error("AddTranslation failed to merge locations")
	}

	// Add with context
	tr3 := &Translation{
		MsgID:           "test",
		Context:         "ctx",
		SourceLocations: []string{"file.go:30"},
	}
	d.AddTranslation(tr3)
	if len(d.ContextTranslations["ctx"]) != 1 {
		t.Error("AddTranslation failed for context")
	}
	
	// Add same ID in same context
	tr4 := &Translation{
		MsgID:           "test",
		Context:         "ctx",
		SourceLocations: []string{"file.go:40"},
	}
	d.AddTranslation(tr4)
	if len(d.ContextTranslations["ctx"]["test"].SourceLocations) != 2 {
		t.Error("AddTranslation failed to merge context locations")
	}
}

func TestDomainMap_AddTranslation(t *testing.T) {
	dm := &DomainMap{}
	dm.AddTranslation("dom1", &Translation{MsgID: "test1"})
	if len(dm.Domains["dom1"].Translations) != 1 {
		t.Error("DomainMap.AddTranslation failed")
	}

	dm.AddTranslation("", &Translation{MsgID: "test_default"})
	if len(dm.Domains["default"].Translations) != 1 {
		t.Error("DomainMap.AddTranslation failed for default domain")
	}
}

func TestDomainMap_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gotext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dm := &DomainMap{}
	dm.AddTranslation("test", &Translation{MsgID: "msg", SourceLocations: []string{"loc:1"}})
	
	err = dm.Save(tmpDir)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	potPath := filepath.Join(tmpDir, "test.pot")
	if _, err := os.Stat(potPath); os.IsNotExist(err) {
		t.Error("pot file was not created")
	}
}

func contains(s, substr string) bool {
	return (len(s) >= len(substr)) && (s[0:len(substr)] == substr || contains(s[1:], substr))
}
