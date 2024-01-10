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

	arabicPo := NewPo()
	arabicPo.ParseFile(arFixture)
	arabic := arabicPo.GetDomain()

	if !arabic.IsTranslated("Load %d more document") {
		t.Error("Arabic singular should be reported as translated.")
	}

	// context
	if !english.IsTranslatedC("One with var: %s", "Ctx") {
		t.Error("Context singular should be reported as translated.")
	}
}

func TestDomain_IsTranslatedN(t *testing.T) {
	englishPo := NewPo()
	englishPo.ParseFile(enUSFixture)
	english := englishPo.GetDomain()

	tests := []struct {
		singular string
		count    int
		expected bool
	}{
		{"Empty plural form singular", 0, false},
		{"Empty plural form singular", 1, true},
		{"Empty plural form singular", 99, false},
		{"One with var: %s", 1, true},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		result := english.IsTranslatedN(tt.singular, tt.count)
		if result != tt.expected {
			t.Errorf("IsTranslatedN() with count %d = %s", tt.count, tt.singular)
		}
	}

	araPo := NewPo()
	araPo.ParseFile(arFixture)
	ara := araPo.GetDomain()

	tests = []struct {
		singular string
		count    int
		expected bool
	}{
		{"Load %d more document", 0, true},
		{"Load %d more document", 1, true},
		{"Load %d more document", 2, true},
		{"Load %d more document", 3, true},
		{"Load %d more document", 10, true},
		{"Load %d more document", 11, true},
		{"Load %d more document", 99, true},
		{"Load %d more document", 100, true},
		{"If n is 0, you should use msgstr[0]", 0, true},
		{"If n is 0, you should use msgstr[0]", 1, false},
		{"If n is 1, you should use msgstr[1]", 1, true},
		{"If n is 1, you should use msgstr[1]", 2, false},
		{"If n is 2, you should use msgstr[2]", 2, true},
		{"If n is 2, you should use msgstr[2]", 3, false},
		{"If n is between 3 and 10, you should use msgstr[3]", 3, true},
		{"If n is between 3 and 10, you should use msgstr[3]", 10, true},
		{"If n is between 3 and 10, you should use msgstr[3]", 11, false},
		{"If n is between 11 and 99, you should use msgstr[4]", 11, true},
		{"If n is between 11 and 99, you should use msgstr[4]", 99, true},
		{"If n is between 11 and 99, you should use msgstr[4]", 100, false},
		{"For any other value of n, msgstr[5] should be used.", 100, true},
		{"For any other value of n, msgstr[5] should be used.", 101, true},
		{"For any other value of n, msgstr[5] should be used.", 312, false},
	}

	for _, tt := range tests {
		result := ara.IsTranslatedN(tt.singular, tt.count)
		if result != tt.expected {
			t.Errorf("IsTranslatedN() with count %d = %s", tt.count, tt.singular)
		}
	}

	if !english.IsTranslatedNC("One with var: %s", 1, "Ctx") {
		t.Error("Context plural should be reported as translated for n=1")
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
