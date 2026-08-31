package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestDomain_AddTranslation_PartialMaps(t *testing.T) {
	t.Run("translations initialized", func(t *testing.T) {
		existing := &Translation{MsgID: "existing"}
		domain := &Domain{
			Translations: TranslationMap{"existing": existing},
		}
		domain.AddTranslation(&Translation{MsgID: "contextual", Context: "ctx"})

		if domain.Translations["existing"] != existing {
			t.Fatal("existing translations were replaced")
		}
		if domain.ContextTranslations["ctx"]["contextual"] == nil {
			t.Fatal("context translations were not initialized")
		}
	})

	t.Run("context translations initialized", func(t *testing.T) {
		existing := &Translation{MsgID: "existing", Context: "ctx"}
		domain := &Domain{
			ContextTranslations: map[string]TranslationMap{
				"ctx": {"existing": existing},
			},
		}
		domain.AddTranslation(&Translation{MsgID: "uncontextualized"})

		if domain.ContextTranslations["ctx"]["existing"] != existing {
			t.Fatal("existing context translations were replaced")
		}
		if domain.Translations["uncontextualized"] == nil {
			t.Fatal("translations were not initialized")
		}
	})
}

func TestDomainMap_AddTranslation_CustomDefaultMergesLocations(t *testing.T) {
	domainMap := &DomainMap{Default: "messages"}
	domainMap.AddTranslation("", &Translation{
		MsgID:           "same",
		SourceLocations: []string{"z.go:4"},
	})
	domainMap.AddTranslation("", &Translation{
		MsgID:           "same",
		SourceLocations: []string{"a.go:2"},
	})

	domain := domainMap.Domains["messages"]
	if domain == nil {
		t.Fatal("custom default domain was not created")
	}
	translation := domain.Translations["same"]
	if translation == nil {
		t.Fatal("translation was not added to the custom default domain")
	}
	if !reflect.DeepEqual(translation.SourceLocations, []string{"z.go:4", "a.go:2"}) {
		t.Fatalf("locations = %v, want source order preserved", translation.SourceLocations)
	}
}

func TestDomain_Dump_PartialContextMapHasNoEmptyEntry(t *testing.T) {
	domain := &Domain{
		ContextTranslations: map[string]TranslationMap{
			"ctx": {"id": {MsgID: "id", Context: "ctx"}},
		},
	}
	dump := domain.Dump()
	if strings.HasPrefix(dump, "\n") || !contains(dump, "msgid \"id\"") {
		t.Fatalf("unexpected partial-domain dump: %q", dump)
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

func FuzzTranslationDumpPreservesState(f *testing.F) {
	f.Add("id", "", "", "z.go:4\na.go:2")
	f.Add("id", "plural", "ctx", "")
	f.Add("", "", "", `location with "quotes"`)

	f.Fuzz(func(t *testing.T, msgID, msgIDPlural, context, locationData string) {
		if len(msgID)+len(msgIDPlural)+len(context)+len(locationData) > 64<<10 {
			return
		}

		var locations []string
		if locationData != "" {
			locations = strings.Split(locationData, "\n")
		}
		before := append([]string(nil), locations...)
		translation := &Translation{
			MsgID:           msgID,
			MsgIDPlural:     msgIDPlural,
			Context:         context,
			SourceLocations: locations,
		}
		firstDump := translation.Dump()
		secondDump := translation.Dump()
		if firstDump != secondDump {
			t.Fatalf("Dump changed across repeated calls: %q != %q", firstDump, secondDump)
		}
		if !reflect.DeepEqual(translation.SourceLocations, before) {
			t.Fatalf("Dump mutated source locations: got %v, want %v", translation.SourceLocations, before)
		}
	})
}

func contains(s, substr string) bool {
	return (len(s) >= len(substr)) && (s[0:len(substr)] == substr || contains(s[1:], substr))
}
