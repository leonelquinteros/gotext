package gotext

import (
	"testing"
)

const (
	enUSFixture = "fixtures/en_US/default.po"
	arFixture   = "fixtures/ar/categories.po"
)

// since both Po and Mo just pass-through to Domain for MarshalBinary and UnmarshalBinary, test it here
func TestBinaryEncoding(t *testing.T) {
	// Create po objects
	po := NewPo()
	po2 := NewPo()

	// Parse file
	po.ParseFile(enUSFixture)

	buff, err := po.GetDomain().MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = po2.GetDomain().UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	// Test translations
	tr := po2.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = po2.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestDomain_GetTranslations(t *testing.T) {
	po := NewPo()
	po.ParseFile(enUSFixture)

	domain := po.GetDomain()
	all := domain.GetTranslations()

	if len(all) != len(domain.translations) {
		t.Error("lengths should match")
	}

	for k, v := range domain.translations {
		if all[k] == v {
			t.Error("GetTranslations should be returning a copy, but pointers are equal")
		}
		if all[k].ID != v.ID {
			t.Error("IDs should match")
		}
		if all[k].PluralID != v.PluralID {
			t.Error("PluralIDs should match")
		}
		if all[k].dirty != v.dirty {
			t.Error("dirty flag should match")
		}
		if len(all[k].Trs) != len(v.Trs) {
			t.Errorf("Trs length does not match: %d != %d", len(all[k].Trs), len(v.Trs))
		}
		if len(all[k].Refs) != len(v.Refs) {
			t.Errorf("Refs length does not match: %d != %d", len(all[k].Refs), len(v.Refs))
		}
	}
}

func TestDomain_GetCtxTranslations(t *testing.T) {
	po := NewPo()
	po.ParseFile(enUSFixture)

	domain := po.GetDomain()
	all := domain.GetCtxTranslations()

	if len(all) != len(domain.contextTranslations) {
		t.Error("lengths should match")
	}

	if domain.contextTranslations["Ctx"] == nil {
		t.Error("Context 'Ctx' should exist")
	}

	for k, v := range domain.contextTranslations {
		for kk, vv := range v {
			if all[k][kk] == vv {
				t.Error("GetCtxTranslations should be returning a copy, but pointers are equal")
			}
			if all[k][kk].ID != vv.ID {
				t.Error("IDs should match")
			}
			if all[k][kk].PluralID != vv.PluralID {
				t.Error("PluralIDs should match")
			}
			if all[k][kk].dirty != vv.dirty {
				t.Error("dirty flag should match")
			}
			if len(all[k][kk].Trs) != len(vv.Trs) {
				t.Errorf("Trs length does not match: %d != %d", len(all[k][kk].Trs), len(vv.Trs))
			}
			if len(all[k][kk].Refs) != len(vv.Refs) {
				t.Errorf("Refs length does not match: %d != %d", len(all[k][kk].Refs), len(vv.Refs))
			}
		}

	}
}

func TestDomain_IsTranslated(t *testing.T) {
	englishPo := NewPo()
	englishPo.ParseFile(enUSFixture)
	english := englishPo.GetDomain()

	// singular and plural
	if english.IsTranslated("My Text") {
		t.Error("'My text' should be reported as translated.")
	}
	if english.IsTranslated("Another string") {
		t.Error("'Another string' should be reported as not translated.")
	}
	if !english.IsTranslatedN("Empty plural form singular", 1) {
		t.Error("'Empty plural form singular' should be reported as translated for n=1.")
	}
	if english.IsTranslatedN("Empty plural form singular", 0) {
		t.Error("'Empty plural form singular' should be reported as not translated for n=0.")
	}

	arabicPo := NewPo()
	arabicPo.ParseFile(arFixture)
	arabic := arabicPo.GetDomain()

	// multiple plurals
	if !arabic.IsTranslated("Load %d more document") {
		t.Error("Arabic singular should be reported as translated.")
	}
	if !arabic.IsTranslatedN("Load %d more document", 0) {
		t.Error("Arabic plural should be reported as translated for n=0.")
	}
	if !arabic.IsTranslatedN("Load %d more document", 1) {
		t.Error("Arabic plural should be reported as translated for n=1.")
	}
	if !arabic.IsTranslatedN("Load %d more document", 100) {
		t.Error("Arabic plural should be reported as translated for n=100.")
	}

	// context
	if !english.IsTranslatedC("One with var: %s", "Ctx") {
		t.Error("Context singular should be reported as translated.")
	}
	if !english.IsTranslatedNC("One with var: %s", 0, "Ctx") {
		t.Error("Context plural should be reported as translated for n=0")
	}
	if !english.IsTranslatedNC("One with var: %s", 2, "Ctx") {
		t.Error("Context plural should be reported as translated for n=2")
	}
}

func TestDomain_CheckExportFormatting(t *testing.T) {
	po := NewPo()
	po.Set("myid", "test string\nwith \"newline\"")
	poBytes, _ := po.MarshalText()

	expectedOutput := `msgid ""
msgstr ""

msgid "myid"
msgstr ""
"test string\n"
"with \"newline\""`

	if string(poBytes) != expectedOutput {
		t.Errorf("Exported PO format does not match. Received:\n\n%v\n\n\nExpected:\n\n%v", string(poBytes), expectedOutput)
	}
}
