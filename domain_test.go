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

func TestDomain_GetWithVar(t *testing.T) {
	po := NewPo()
	po.ParseFile(enUSFixture)

	domain := po.GetDomain()

	// Test singular with variable
	v := "My Text"
	tr := domain.Get(v)
	if tr != "My Text" {
		t.Errorf("Expected 'MyText' but got '%s'", tr)
	}

	tr = po.Get(v)
	if tr != "My Text" {
		t.Errorf("Expected 'MyText' but got '%s'", tr)
	}
}

func TestDomain_Append(t *testing.T) {
	d := NewDomain()
	d.Set("test", "translated")

	b := []byte("prefix: ")
	res := d.Append(b, "test")
	if string(res) != "prefix: translated" {
		t.Errorf("Expected 'prefix: translated', got '%s'", string(res))
	}

	res = d.Append(nil, "missing")
	if string(res) != "missing" {
		t.Errorf("Expected 'missing', got '%s'", string(res))
	}
}

func TestDomain_AppendN(t *testing.T) {
	d := NewDomain()
	d.SetN("one", "many", 1, "singular")
	d.SetN("one", "many", 2, "plural")

	res := d.AppendN(nil, "one", "many", 1)
	if string(res) != "singular" {
		t.Errorf("Expected 'singular', got '%s'", string(res))
	}

	res = d.AppendN(nil, "one", "many", 2)
	if string(res) != "plural" {
		t.Errorf("Expected 'plural', got '%s'", string(res))
	}

	res = d.AppendN(nil, "missing", "missings", 1)
	if string(res) != "missing" {
		t.Errorf("Expected 'missing', got '%s'", string(res))
	}

	res = d.AppendN(nil, "missing", "missings", 2)
	if string(res) != "missings" {
		t.Errorf("Expected 'missings', got '%s'", string(res))
	}
}

func TestDomain_AppendC(t *testing.T) {
	d := NewDomain()
	d.SetC("test", "ctx", "translated")

	res := d.AppendC(nil, "test", "ctx")
	if string(res) != "translated" {
		t.Errorf("Expected 'translated', got '%s'", string(res))
	}

	res = d.AppendC(nil, "test", "wrong_ctx")
	if string(res) != "test" {
		t.Errorf("Expected 'test', got '%s'", string(res))
	}
}

func TestDomain_AppendNC(t *testing.T) {
	d := NewDomain()
	d.SetNC("one", "many", "ctx", 1, "singular")
	d.SetNC("one", "many", "ctx", 2, "plural")

	res := d.AppendNC(nil, "one", "many", 1, "ctx")
	if string(res) != "singular" {
		t.Errorf("Expected 'singular', got '%s'", string(res))
	}

	res = d.AppendNC(nil, "one", "many", 2, "ctx")
	if string(res) != "plural" {
		t.Errorf("Expected 'plural', got '%s'", string(res))
	}
}

func TestDomain_SetNC(t *testing.T) {
	d := NewDomain()
	d.SetNC("one", "many", "ctx", 1, "singular")
	// Update existing
	d.SetNC("one", "many", "ctx", 1, "singular_updated")

	res := d.GetNC("one", "many", 1, "ctx")
	if res != "singular_updated" {
		t.Errorf("Expected 'singular_updated', got '%s'", res)
	}

	// New one in existing context
	d.SetNC("two", "plural_two", "ctx", 1, "two_singular")
	res = d.GetNC("two", "plural_two", 1, "ctx")
	if res != "two_singular" {
		t.Errorf("Expected 'two_singular', got '%s'", res)
	}
}

func TestDomain_Refs(t *testing.T) {
	d := NewDomain()
	refs := []string{"file.go:10", "file.go:20"}
	d.SetRefs("test", refs)

	gotRefs := d.GetRefs("test")
	if len(gotRefs) != 2 || gotRefs[0] != refs[0] || gotRefs[1] != refs[1] {
		t.Errorf("Expected %v, got %v", refs, gotRefs)
	}

	if d.GetRefs("missing") != nil {
		t.Error("Expected nil for missing refs")
	}

	// Update refs
	newRefs := []string{"file.go:30"}
	d.SetRefs("test", newRefs)
	gotRefs = d.GetRefs("test")
	if len(gotRefs) != 1 || gotRefs[0] != newRefs[0] {
		t.Errorf("Expected %v, got %v", newRefs, gotRefs)
	}
}

func TestDomain_HeaderMap(t *testing.T) {
	d := NewDomain()
	d.Headers.Del("Missing") // No-op but increases coverage

	d.Headers.Set("Key", "Value")
	if d.Headers.Get("Key") != "Value" {
		t.Error("Header Get/Set failed")
	}

	d.Headers.Add("Key", "Value2")
	values := d.Headers.Values("Key")
	if len(values) != 2 || values[1] != "Value2" {
		t.Error("Header Add failed")
	}

	d.Headers.Del("Key")
	if d.Headers.Get("Key") != "" {
		t.Error("Header Del failed")
	}

	var nilHeaders HeaderMap
	if nilHeaders.Get("Any") != "" {
		t.Error("Nil headers should return empty string")
	}
	if nilHeaders.Values("Any") != nil {
		t.Error("Nil headers should return nil values")
	}
}

func TestDomain_SetPluralResolver(t *testing.T) {
	d := NewDomain()
	d.SetPluralResolver(func(n int) int {
		return 5
	})
	if d.pluralForm(10) != 5 {
		t.Error("Custom plural resolver failed")
	}
}
