package gotext

import (
	"bytes"
	"encoding/gob"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	enUSFixture = "fixtures/en_US/default.po"
	arFixture   = "fixtures/ar/categories.po"
)

func encodeTestGob(t testing.TB, value any) []byte {
	t.Helper()

	var encoded bytes.Buffer
	if err := gob.NewEncoder(&encoded).Encode(value); err != nil {
		t.Fatalf("encode test gob: %v", err)
	}
	return encoded.Bytes()
}

// since both Po and Mo just pass-through to Domain for MarshalBinary and UnmarshalBinary, test it here
func TestBinaryEncoding(t *testing.T) {
	data := encodeTestGob(t, &TranslatorEncoding{
		Language: "en_US",
		Translations: map[string]*Translation{
			"My text": {
				ID:  "My text",
				Trs: map[int]string{0: "Translated text"},
			},
			"language": {
				ID:  "language",
				Trs: map[int]string{0: "en_US"},
			},
		},
	})

	po := NewPo()
	if err := po.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if got := po.Get("My text"); got != "Translated text" {
		t.Errorf("decoded My text = %q, want %q", got, "Translated text")
	}
	if got := po.Get("language"); got != "en_US" {
		t.Errorf("decoded language = %q, want %q", got, "en_US")
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

func TestDomain_TranslationCopiesOwnPluralMap(t *testing.T) {
	d := NewDomain()

	translation := NewTranslation()
	translation.ID = "id"
	translation.Trs[2] = "plural"
	d.translations[translation.ID] = translation

	contextTranslation := NewTranslation()
	contextTranslation.ID = "contextual id"
	contextTranslation.Trs[1] = "contextual plural"
	d.contextTranslations["context"] = map[string]*Translation{
		contextTranslation.ID: contextTranslation,
	}

	copied := d.GetTranslations()[translation.ID]
	copied.Trs[2] = "changed"
	if got := d.translations[translation.ID].Trs[2]; got != "plural" {
		t.Fatalf("GetTranslations should own its Trs map, got %q in source", got)
	}

	copiedContext := d.GetCtxTranslations()["context"][contextTranslation.ID]
	copiedContext.Trs[1] = "changed"
	if got := d.contextTranslations["context"][contextTranslation.ID].Trs[1]; got != "contextual plural" {
		t.Fatalf("GetCtxTranslations should own its Trs map, got %q in source", got)
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

	d.Headers.Set("Key", "Value")
	if d.Headers.Get("Key") != "Value" {
		t.Error("Header Get/Set failed")
	}
	d.Headers.Del("Missing")
	if got := d.Headers.Get("Key"); got != "Value" {
		t.Errorf("Del missing header changed Key: got %q, want Value", got)
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

	po := NewPo()
	po.SetPluralResolver(func(n int) int {
		return 6
	})
	if got := po.GetDomain().pluralForm(10); got != 6 {
		t.Errorf("Po custom plural resolver = %d, want 6", got)
	}
}

func TestDomain_CustomPluralResolverCanReenter(t *testing.T) {
	d := NewDomain()
	d.SetPluralResolver(func(int) int {
		d.Set("reentrant", "value")
		return 0
	})

	done := make(chan struct{}, 1)
	go func() {
		d.SetN("message", "messages", 2, "message")
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("custom plural resolver deadlocked while reentering Domain")
	}

	if got := d.Get("reentrant"); got != "value" {
		t.Fatalf("reentrant resolver mutation produced %q, want %q", got, "value")
	}
}

func TestDomain_MarshalPluralEscapingAndOrder(t *testing.T) {
	d := NewDomain()
	trans := NewTranslation()
	trans.ID = "\"leading\" and embedded \"quote\" \\raw"
	trans.PluralID = "\"plural leading\nembedded \"quote\" and \\\"preserved with \\raw"
	trans.Trs[2] = "line two\nwith \"quote\" and \\raw"
	trans.Trs[0] = "\"translation leading\nwith \\raw"
	trans.Trs[1] = `already \"escaped`
	d.translations[trans.ID] = trans

	data, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	output := string(data)
	previous := strings.Index(output, "msgid_plural")
	if previous == -1 {
		t.Fatal("expected a plural message")
	}
	for _, marker := range []string{"msgstr[0]", "msgstr[1]", "msgstr[2]"} {
		position := strings.Index(output, marker)
		if position <= previous {
			t.Fatalf("expected %s after the preceding plural field in %q", marker, output)
		}
		previous = position
	}

	for _, fragment := range []string{
		`msgid "\"leading\" and embedded \"quote\" \\raw"`,
		`"embedded \"quote\" and \"preserved with \\raw"`,
		`msgstr[1] "already \"escaped"`,
		`"with \"quote\" and \\raw`,
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("MarshalText output missing escaped fragment %q:\n%s", fragment, output)
		}
	}
}

func TestDomain_ParseHeadersUsesSplitSeqAndCut(t *testing.T) {
	d := NewDomain()
	header := NewTranslation()
	header.Trs[0] = "\n\nlanguage: en_US:UTF-8\nLANGUAGE: en_GB\nX-Custom: value=with=equals\nX-Custom: duplicate\npLuRaL-fOrMs: nplurals=3; malformed; plural=n != 1; ignored=a=b\n"
	d.translations[""] = header

	d.parseHeaders()

	if got := d.Language; got != "en_GB" {
		t.Errorf("Language = %q, want %q", got, "en_GB")
	}
	if got := d.Headers.Get("language"); got != "en_US:UTF-8" {
		t.Errorf("lower-case Language header = %q, want %q", got, "en_US:UTF-8")
	}
	if got := d.Headers.Get("LANGUAGE"); got != "en_GB" {
		t.Errorf("upper-case Language header = %q, want %q", got, "en_GB")
	}
	customValues := d.Headers.Values("X-Custom")
	if len(customValues) != 2 || customValues[0] != "value=with=equals" || customValues[1] != "duplicate" {
		t.Errorf("duplicate custom headers = %v", customValues)
	}

	wantPluralForms := "nplurals=3; malformed; plural=n != 1; ignored=a=b"
	if d.PluralForms != wantPluralForms {
		t.Errorf("PluralForms = %q, want %q", d.PluralForms, wantPluralForms)
	}
	if d.nplurals != 3 {
		t.Errorf("nplurals = %d, want 3", d.nplurals)
	}
	if d.plural != "n != 1" {
		t.Errorf("plural expression = %q, want %q", d.plural, "n != 1")
	}
	if got := d.pluralForm(1); got != 0 {
		t.Errorf("plural form for one = %d, want 0", got)
	}
	if got := d.pluralForm(2); got != 1 {
		t.Errorf("plural form for two = %d, want 1", got)
	}
}

func TestDomain_SetRefsOwnsSliceAndRefreshesExisting(t *testing.T) {
	d := NewDomain()
	trans := NewTranslation()
	trans.ID = "message"
	d.translations[trans.ID] = trans

	refs := []string{"file.go:1"}
	d.SetRefs(trans.ID, refs)
	refs[0] = "file.go:2"

	if got := d.GetRefs(trans.ID); len(got) != 1 || got[0] != "file.go:1" {
		t.Fatalf("SetRefs should own its input, got %v", got)
	}
	if trans.IsStale() {
		t.Fatal("SetRefs should mark an existing translation dirty")
	}
	d.DropStaleTranslations()
	if _, ok := d.translations[trans.ID]; !ok {
		t.Fatal("DropStaleTranslations removed a translation refreshed by SetRefs")
	}

	got := d.GetRefs(trans.ID)
	got[0] = "file.go:3"
	if got = d.GetRefs(trans.ID); got[0] != "file.go:1" {
		t.Fatalf("GetRefs should return an independent slice, got %v", got)
	}
}

func TestDomain_BinarySnapshotAndInvalidDecodeAreSafe(t *testing.T) {
	d := NewDomain()
	d.Headers.Set("X-Test", "before")
	d.Set("message", "before")
	d.SetRefs("message", []string{"file.go:1"})

	data, err := d.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	d.Headers.Set("X-Test", "after")
	d.Set("message", "after")

	restored := NewDomain()
	if err := restored.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}
	if got := restored.Get("message"); got != "before" {
		t.Fatalf("binary snapshot message = %q, want %q", got, "before")
	}
	if got := restored.Headers.Get("X-Test"); got != "before" {
		t.Fatalf("binary snapshot header = %q, want %q", got, "before")
	}
	if got := restored.GetRefs("message"); len(got) != 1 || got[0] != "file.go:1" {
		t.Fatalf("binary snapshot refs = %v", got)
	}

	if err := d.UnmarshalBinary([]byte("not gob")); err == nil {
		t.Fatal("invalid binary should return an error")
	}
	if got := d.Get("message"); got != "after" {
		t.Fatalf("invalid decode changed translation to %q", got)
	}
	if got := d.Headers.Get("X-Test"); got != "after" {
		t.Fatalf("invalid decode changed header to %q", got)
	}
}

func TestDomain_UnmarshalBinaryInvalidPluralUsesFallback(t *testing.T) {
	var encoded bytes.Buffer
	if err := gob.NewEncoder(&encoded).Encode(&TranslatorEncoding{
		Plural: "n +",
	}); err != nil {
		t.Fatal(err)
	}

	d := NewDomain()
	d.SetPluralResolver(func(int) int {
		return 7
	})
	if err := d.UnmarshalBinary(encoded.Bytes()); err != nil {
		t.Fatal(err)
	}
	if got := d.pluralForm(2); got != 7 {
		t.Fatalf("invalid plural should use custom fallback, got %d", got)
	}

	d.SetPluralResolver(nil)
	if got := d.pluralForm(1); got != 0 {
		t.Fatalf("invalid plural singular fallback = %d, want 0", got)
	}
	if got := d.pluralForm(2); got != 1 {
		t.Fatalf("invalid plural plural fallback = %d, want 1", got)
	}
}

func TestDomainConcurrentSupportedAccess(t *testing.T) {
	d := NewDomain()
	d.Set("message", "initial")
	d.SetC("message", "context", "initial context")
	d.SetRefs("message", []string{"file.go:1"})
	seed, err := d.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	const iterations = 128
	var wg sync.WaitGroup
	wg.Add(8)

	go func() {
		defer wg.Done()
		for range iterations {
			d.Set("message", "value")
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			d.SetC("message", "context", "context value")
		}
	}()
	go func() {
		defer wg.Done()
		for i := range iterations {
			d.SetN("message", "messages", i, "plural value")
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			d.SetRefs("message", []string{"file.go:1", "file.go:2"})
		}
	}()
	go func() {
		defer wg.Done()
		for i := range iterations {
			if i%2 == 0 {
				d.SetPluralResolver(func(int) int { return 0 })
			} else {
				d.SetPluralResolver(func(int) int { return 1 })
			}
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			refs := d.GetRefs("message")
			if len(refs) > 0 {
				refs[0] = "caller mutation"
			}
			for _, trans := range d.GetTranslations() {
				if trans != nil {
					trans.Trs[0] = "caller mutation"
				}
			}
			for _, translations := range d.GetCtxTranslations() {
				for _, trans := range translations {
					if trans != nil {
						trans.Trs[0] = "caller mutation"
					}
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			if _, err := d.MarshalText(); err != nil {
				t.Errorf("MarshalText: %v", err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			if _, err := d.MarshalBinary(); err != nil {
				t.Errorf("MarshalBinary: %v", err)
			}
			if err := d.UnmarshalBinary(seed); err != nil {
				t.Errorf("UnmarshalBinary: %v", err)
			}
		}
	}()

	wg.Wait()
}
