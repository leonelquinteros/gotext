package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTranslation_AddLocations(t *testing.T) {
	tr := &Translation{MsgID: "test"}
	first := []string{"file1.go:10"}
	tr.AddLocations(first)
	first[0] = "caller mutated its input"
	if !reflect.DeepEqual(tr.SourceLocations, []string{"file1.go:10"}) {
		t.Fatalf("locations after first append = %v, want copied input", tr.SourceLocations)
	}

	tr.AddLocations([]string{"file2.go:20"})
	if !reflect.DeepEqual(tr.SourceLocations, []string{"file1.go:10", "file2.go:20"}) {
		t.Fatalf("locations after append = %v, want source order preserved", tr.SourceLocations)
	}
	before := append([]string(nil), tr.SourceLocations...)
	tr.AddLocations(nil)
	if !reflect.DeepEqual(tr.SourceLocations, before) {
		t.Fatalf("nil append changed locations: got %v, want %v", tr.SourceLocations, before)
	}
}

func TestTranslation_Dump(t *testing.T) {
	tests := []struct {
		name        string
		translation Translation
		want        string
	}{
		{
			name:        "sorted references and escaped message",
			translation: Translation{MsgID: "line \"quoted\"\nnext", SourceLocations: []string{"z.go:4", "a.go:2"}},
			want: `#: a.go:2
#: z.go:4
msgid "line \"quoted\"\nnext"
msgstr ""`,
		},
		{
			name:        "plural",
			translation: Translation{MsgID: "test", MsgIDPlural: "tests"},
			want: `msgid "test"
msgid_plural "tests"
msgstr[0] ""
msgstr[1] ""`,
		},
		{
			name:        "context",
			translation: Translation{MsgID: "test", Context: "ctx"},
			want: `msgctxt ctx
msgid "test"
msgstr ""`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeLocations := append([]string(nil), test.translation.SourceLocations...)
			firstDump := test.translation.Dump()
			if firstDump != test.want {
				t.Fatalf("Dump() = %q, want %q", firstDump, test.want)
			}
			if secondDump := test.translation.Dump(); secondDump != firstDump {
				t.Fatalf("Dump() changed across calls: %q != %q", secondDump, firstDump)
			}
			if !reflect.DeepEqual(test.translation.SourceLocations, beforeLocations) {
				t.Fatalf("Dump() mutated source locations: got %v, want %v",
					test.translation.SourceLocations, beforeLocations)
			}
		})
	}

	var nilTranslation *Translation
	if got := nilTranslation.Dump(); got != "" {
		t.Fatalf("nil Translation.Dump() = %q, want empty output", got)
	}
}

func TestDomain_AddTranslation(t *testing.T) {
	domain := &Domain{}
	domain.AddTranslation(&Translation{
		MsgID:           "test",
		SourceLocations: []string{"file.go:10"},
	})
	domain.AddTranslation(&Translation{
		MsgID:           "test",
		SourceLocations: []string{"file.go:20"},
	})
	translation := domain.Translations["test"]
	if translation == nil {
		t.Fatal("uncontextualized translation was not added")
	}
	if !reflect.DeepEqual(translation.SourceLocations, []string{"file.go:10", "file.go:20"}) {
		t.Fatalf("uncontextualized locations = %v, want merged source order", translation.SourceLocations)
	}

	domain.AddTranslation(&Translation{
		MsgID:           "test",
		Context:         "ctx",
		SourceLocations: []string{"file.go:30"},
	})
	domain.AddTranslation(&Translation{
		MsgID:           "test",
		Context:         "ctx",
		SourceLocations: []string{"file.go:40"},
	})
	contextual := domain.ContextTranslations["ctx"]["test"]
	if contextual == nil {
		t.Fatal("contextual translation was not added")
	}
	if contextual.MsgID != "test" || contextual.Context != "ctx" ||
		!reflect.DeepEqual(contextual.SourceLocations, []string{"file.go:30", "file.go:40"}) {
		t.Fatalf("contextual translation = %#v, want merged context metadata", contextual)
	}
	if len(domain.Translations) != 1 || len(domain.ContextTranslations) != 1 {
		t.Fatalf("domain maps = %#v/%#v, want one entry in each map",
			domain.Translations, domain.ContextTranslations)
	}
}

func TestDomain_AddTranslation_PartialMaps(t *testing.T) {
	t.Run("contextual insertion preserves initialized translations", func(t *testing.T) {
		existing := &Translation{MsgID: "existing"}
		domain := &Domain{Translations: TranslationMap{"existing": existing}}
		domain.AddTranslation(&Translation{MsgID: "contextual", Context: "ctx"})

		if domain.Translations["existing"] != existing {
			t.Fatal("existing translations were replaced")
		}
		translation := domain.ContextTranslations["ctx"]["contextual"]
		if translation == nil || translation.MsgID != "contextual" || translation.Context != "ctx" {
			t.Fatalf("contextual insertion = %#v, want initialized translation", translation)
		}
	})

	t.Run("uncontextualized insertion preserves initialized contexts", func(t *testing.T) {
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
		translation := domain.Translations["uncontextualized"]
		if translation == nil || translation.MsgID != "uncontextualized" || translation.Context != "" {
			t.Fatalf("uncontextualized insertion = %#v, want initialized translation", translation)
		}
	})

	t.Run("nil context bucket is initialized", func(t *testing.T) {
		domain := &Domain{
			ContextTranslations: map[string]TranslationMap{"ctx": nil},
		}
		domain.AddTranslation(&Translation{MsgID: "id", Context: "ctx"})

		translation := domain.ContextTranslations["ctx"]["id"]
		if translation == nil || translation.MsgID != "id" || translation.Context != "ctx" {
			t.Fatalf("nil context bucket insertion = %#v, want translation", translation)
		}
	})

	t.Run("nil existing entry is replaced", func(t *testing.T) {
		domain := &Domain{Translations: TranslationMap{"id": nil}}
		domain.AddTranslation(&Translation{MsgID: "id"})

		translation := domain.Translations["id"]
		if translation == nil || translation.MsgID != "id" {
			t.Fatalf("nil existing entry = %#v, want replacement translation", translation)
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

func TestDomain_Dump_PartialMaps(t *testing.T) {
	domain := &Domain{
		Translations: TranslationMap{"ignored": nil},
		ContextTranslations: map[string]TranslationMap{
			"empty":     nil,
			"ctx":       {"id": {MsgID: "id", Context: "ctx", SourceLocations: []string{"z.go:2", "a.go:1"}}},
			"nil-entry": {"ignored": nil},
		},
	}
	beforeLocations := append([]string(nil), domain.ContextTranslations["ctx"]["id"].SourceLocations...)
	want := `#: a.go:1
#: z.go:2
msgctxt ctx
msgid "id"
msgstr ""`

	firstDump := domain.Dump()
	if firstDump != want {
		t.Fatalf("Dump() = %q, want %q", firstDump, want)
	}
	if secondDump := domain.Dump(); secondDump != firstDump {
		t.Fatalf("Dump() changed across calls: %q != %q", secondDump, firstDump)
	}
	if !reflect.DeepEqual(domain.ContextTranslations["ctx"]["id"].SourceLocations, beforeLocations) {
		t.Fatalf("Dump() mutated nested source locations: got %v, want %v",
			domain.ContextTranslations["ctx"]["id"].SourceLocations, beforeLocations)
	}
	if strings.HasPrefix(firstDump, "\n") || strings.HasSuffix(firstDump, "\n\n") {
		t.Fatalf("partial maps introduced empty dump entries: %q", firstDump)
	}
}

func TestDomainMap_AddTranslation(t *testing.T) {
	domainMap := &DomainMap{}
	domainMap.AddTranslation("dom1", &Translation{MsgID: "test1"})
	domainMap.AddTranslation("", &Translation{MsgID: "test_default"})

	if domainMap.Default != "default" {
		t.Fatalf("Default = %q, want default", domainMap.Default)
	}
	domain := domainMap.Domains["dom1"]
	if domain == nil || domain.Translations["test1"] == nil {
		t.Fatal("named domain translation was not added")
	}
	defaultDomain := domainMap.Domains["default"]
	if defaultDomain == nil {
		t.Fatal("default domain was not created")
	}
	translation := defaultDomain.Translations["test_default"]
	if translation == nil || translation.MsgID != "test_default" || translation.Context != "" {
		t.Fatalf("default translation = %#v, want literal metadata", translation)
	}
	if len(domainMap.Domains) != 2 {
		t.Fatalf("got %d domains, want two", len(domainMap.Domains))
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

	domainMap := &DomainMap{}
	domainMap.AddTranslation("test", &Translation{
		MsgID:           "msg",
		SourceLocations: []string{"loc:1"},
	})

	if err := domainMap.Save(tmpDir); err != nil {
		t.Fatal(err)
	}

	potPath := filepath.Join(tmpDir, "test.pot")
	data, err := os.ReadFile(potPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, expected := range []string{
		`msgid ""`,
		`"Content-Type: text/plain; charset=UTF-8\n"`,
		"#: loc:1",
		`msgid "msg"`,
		`msgstr ""`,
	} {
		if !strings.Contains(content, expected) {
			t.Errorf("saved domain is missing %q: %q", expected, content)
		}
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
		translation := &Translation{
			MsgID:           msgID,
			MsgIDPlural:     msgIDPlural,
			Context:         context,
			SourceLocations: locations,
		}
		before := *translation
		before.SourceLocations = append([]string(nil), locations...)

		firstDump := translation.Dump()
		if firstDump == "" {
			t.Fatal("Dump returned empty output for a non-nil translation")
		}
		if secondDump := translation.Dump(); secondDump != firstDump {
			t.Fatalf("Dump changed across calls: %q != %q", secondDump, firstDump)
		}
		if !reflect.DeepEqual(*translation, before) {
			t.Fatalf("Dump mutated translation: got %#v, want %#v", *translation, before)
		}
	})
}
