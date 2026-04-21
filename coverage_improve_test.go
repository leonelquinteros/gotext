package gotext

import (
	"testing"
)

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

func TestHelper_Appendf(t *testing.T) {
	b := []byte("test: ")
	res := Appendf(b, "Hello %s", "World")
	if string(res) != "test: Hello World" {
		t.Errorf("Expected 'test: Hello World', got '%s'", string(res))
	}
}

func TestTranslation_NewTranslationWithRefs(t *testing.T) {
	refs := []string{"a.go:1", "b.go:2"}
	tr := NewTranslationWithRefs(refs)
	if len(tr.Refs) != 2 || tr.Refs[0] != "a.go:1" {
		t.Error("NewTranslationWithRefs failed")
	}
}

func TestTranslation_IsTranslated(t *testing.T) {
	tr := NewTranslation()
	if tr.IsTranslated() {
		t.Error("Expected false for empty translation")
	}
}

func TestDomain_GetD_Missing(t *testing.T) {
	// Covering the case where Domains map is nil or domain is missing
	l := NewLocale("path", "en")
	res := l.GetD("missing", "test")
	if res != "test" {
		t.Errorf("Expected 'test', got '%s'", res)
	}

	res = l.GetND("missing", "one", "many", 1)
	if res != "one" {
		t.Errorf("Expected 'one', got '%s'", res)
	}
	res = l.GetND("missing", "one", "many", 2)
	if res != "many" {
		t.Errorf("Expected 'many', got '%s'", res)
	}

	res = l.GetDC("missing", "test", "ctx")
	if res != "test" {
		t.Errorf("Expected 'test', got '%s'", res)
	}

	res = l.GetNDC("missing", "one", "many", 1, "ctx")
	if res != "one" {
		t.Errorf("Expected 'one', got '%s'", res)
	}
	res = l.GetNDC("missing", "one", "many", 2, "ctx")
	if res != "many" {
		t.Errorf("Expected 'many', got '%s'", res)
	}
}

func TestPo_MissingWrappers(t *testing.T) {
	po := NewPo()
	// Coverage for SetRefs, GetRefs, SetPluralResolver, etc in Po
	po.SetRefs("id", []string{"ref"})
	refs := po.GetRefs("id")
	if len(refs) != 1 || refs[0] != "ref" {
		t.Error("Po.SetRefs/GetRefs failed")
	}

	po.SetPluralResolver(func(n int) int { return 1 })

	res := po.Append(nil, "test")
	if string(res) != "test" {
		t.Error("Po.Append failed")
	}

	po.SetN("one", "many", 1, "singular")
	res = po.AppendN(nil, "one", "many", 1)
	if string(res) != "singular" {
		t.Error("Po.AppendN failed")
	}

	po.SetC("id", "ctx", "val")
	res = po.AppendC(nil, "id", "ctx")
	if string(res) != "val" {
		t.Error("Po.AppendC failed")
	}

	po.SetNC("id", "plural", "ctx", 1, "val_nc")
	res = po.AppendNC(nil, "id", "plural", 1, "ctx")
	if string(res) != "val_nc" {
		t.Error("Po.AppendNC failed")
	}

	if po.IsTranslated("missing") {
		t.Error("Po.IsTranslated failed")
	}
	if po.IsTranslatedN("missing", 1) {
		t.Error("Po.IsTranslatedN failed")
	}
	if po.IsTranslatedC("missing", "ctx") {
		t.Error("Po.IsTranslatedC failed")
	}
	if po.IsTranslatedNC("missing", 1, "ctx") {
		t.Error("Po.IsTranslatedNC failed")
	}
}

func TestMo_MissingWrappers(t *testing.T) {
	mo := NewMo()
	res := mo.Append(nil, "test")
	if string(res) != "test" {
		t.Error("Mo.Append failed")
	}

	res = mo.AppendN(nil, "one", "many", 1)
	if string(res) != "one" {
		t.Error("Mo.AppendN failed")
	}

	res = mo.AppendC(nil, "id", "ctx")
	if string(res) != "id" {
		t.Error("Mo.AppendC failed")
	}

	res = mo.AppendNC(nil, "id", "plural", 1, "ctx")
	if string(res) != "id" {
		t.Error("Mo.AppendNC failed")
	}

	if mo.IsTranslated("id") {
		t.Error("Mo.IsTranslated failed")
	}
	if mo.IsTranslatedN("id", 1) {
		t.Error("Mo.IsTranslatedN failed")
	}
	if mo.IsTranslatedC("id", "ctx") {
		t.Error("Mo.IsTranslatedC failed")
	}
	if mo.IsTranslatedNC("id", 1, "ctx") {
		t.Error("Mo.IsTranslatedNC failed")
	}
}

func TestLocale_MissingIsTranslatedWrappers(t *testing.T) {
	l := NewLocale("path", "en")
	if l.IsTranslated("test") {
		t.Error("Expected false for missing domain")
	}
	if l.IsTranslatedN("test", 1) {
		t.Error("Expected false for missing domain")
	}
	if l.IsTranslatedC("test", "ctx") {
		t.Error("Expected false for missing domain")
	}
	if l.IsTranslatedNC("test", 1, "ctx") {
		t.Error("Expected false for missing domain")
	}
}

func TestGotext_MissingWrappers(t *testing.T) {
	// These are just wrappers around the global config
	if IsTranslatedD("missing", "id") {
		t.Error("IsTranslatedD failed")
	}
	if IsTranslatedDC("missing", "id", "ctx") {
		t.Error("IsTranslatedDC failed")
	}
	
	if GetStorage() == nil {
		t.Error("GetStorage should not be nil")
	}
	
	SetStorage(GetStorage()) // Coverage
	
	Configure("fixtures", "en_US", "default")
	if GetD("default", "My text") != translatedText {
		t.Error("GetD failed")
	}
}
