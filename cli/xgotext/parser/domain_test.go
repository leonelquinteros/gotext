package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func createTestSymlink(t *testing.T, target, link string) {
	t.Helper()
	if runtime.GOOS == "js" || runtime.GOOS == "plan9" || runtime.GOOS == "wasip1" {
		t.Skipf("symbolic links are unavailable on %s", runtime.GOOS)
	}
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" && os.IsPermission(err) {
			t.Skipf("symbolic links require privileges on %s: %v", runtime.GOOS, err)
		}
		t.Fatalf("create symlink %q -> %q: %v", link, target, err)
	}
}

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
			want: `msgctxt "ctx"
msgid "test"
msgstr ""`,
		},
		{
			name: "escaped multiline context",
			translation: Translation{
				MsgID:      "test",
				Context:    "ctx \"quoted\"\\slash\nnext",
				HasContext: true,
			},
			want: `msgctxt "ctx \"quoted\"\\slash\nnext"
msgid "test"
msgstr ""`,
		},
		{
			name:        "explicit empty context",
			translation: Translation{MsgID: "test", HasContext: true},
			want: `msgctxt ""
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

func TestDomain_AddTranslation_ExplicitEmptyContext(t *testing.T) {
	domain := &Domain{}
	domain.AddTranslation(&Translation{
		MsgID:           "test",
		HasContext:      true,
		SourceLocations: []string{"file.go:10"},
	})
	domain.AddTranslation(&Translation{
		MsgID:           "test",
		HasContext:      true,
		SourceLocations: []string{"file.go:20"},
	})

	translation := domain.ContextTranslations[""]["test"]
	if translation == nil {
		t.Fatal("explicit-empty contextual translation was not added")
	}
	if !translation.HasContext || translation.Context != "" ||
		!reflect.DeepEqual(translation.SourceLocations, []string{"file.go:10", "file.go:20"}) {
		t.Fatalf("explicit-empty contextual translation = %#v, want merged context metadata", translation)
	}
	if _, ok := domain.Translations["test"]; ok {
		t.Fatal("explicit-empty contextual translation was added as uncontextualized")
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
msgctxt "ctx"
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
	tmpDir := t.TempDir()

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

func TestDomain_SaveUsesExplicitPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "explicit.pot")
	domain := &Domain{}
	domain.AddTranslation(&Translation{MsgID: "explicit"})

	if err := domain.Save(path); err != nil {
		t.Fatalf("Save(%q) failed: %v", path, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read explicitly saved domain: %v", err)
	}
	if !strings.Contains(string(data), `msgid "explicit"`) {
		t.Fatalf("explicitly saved domain is missing translation: %q", data)
	}
}

func TestDomainMap_SaveRejectsLexicalTraversal(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "output")
	outside := filepath.Join(parent, "escaped.pot")
	sentinel := []byte("do not overwrite")
	if err := os.WriteFile(outside, sentinel, 0o600); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	domainMap := &DomainMap{}
	domainMap.AddTranslation("../escaped", &Translation{MsgID: "replacement"})
	if err := domainMap.Save(directory); err == nil {
		t.Fatal("Save accepted a domain name that escapes the output root")
	}

	if got, err := os.ReadFile(outside); err != nil {
		t.Fatalf("read sentinel: %v", err)
	} else if string(got) != string(sentinel) {
		t.Fatalf("sentinel after traversal attempt = %q, want %q", got, sentinel)
	}
}

func TestDomainMap_SaveRejectsEscapingSymlink(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "output")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("create output directory: %v", err)
	}
	outside := filepath.Join(parent, "outside.pot")
	sentinel := []byte("do not overwrite")
	if err := os.WriteFile(outside, sentinel, 0o600); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	createTestSymlink(t, "../outside.pot", filepath.Join(directory, "linked.pot"))

	domainMap := &DomainMap{}
	domainMap.AddTranslation("linked", &Translation{MsgID: "replacement"})
	if err := domainMap.Save(directory); err == nil {
		t.Fatal("Save accepted a symlink that escapes the output root")
	}

	if got, err := os.ReadFile(outside); err != nil {
		t.Fatalf("read sentinel: %v", err)
	} else if string(got) != string(sentinel) {
		t.Fatalf("sentinel after symlink attempt = %q, want %q", got, sentinel)
	}
}

func TestDomainMap_SaveAllowsContainedSymlink(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "output")
	nested := filepath.Join(directory, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested output directory: %v", err)
	}
	target := filepath.Join(nested, "inside.pot")
	createTestSymlink(t, "inside.pot", filepath.Join(nested, "linked.pot"))

	domainMap := &DomainMap{}
	domainMap.AddTranslation("nested/linked", &Translation{MsgID: "contained"})
	if err := domainMap.Save(directory); err != nil {
		t.Fatalf("Save rejected a contained symlink: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read contained symlink target: %v", err)
	}
	if !strings.Contains(string(data), `msgid "contained"`) {
		t.Fatalf("contained symlink target is missing translation: %q", data)
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
