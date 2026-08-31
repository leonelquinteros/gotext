/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"bytes"
	"embed"
	"encoding/gob"
	"errors"
	"os"
	"path"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

type reentrantTranslator struct {
	locale     *Locale
	po         *Po
	marshalErr error
	calls      map[string]int
}

func newReentrantTranslator(locale *Locale) *reentrantTranslator {
	po := NewPo()
	po.Set("message", "translated")
	po.SetN("message", "messages", 2, "plural translated")
	po.SetC("message", "context", "context translated")
	po.SetNC("message", "messages", "context", 2, "context plural translated")
	return &reentrantTranslator{locale: locale, po: po, calls: make(map[string]int)}
}
func waitForLocaleCall(t *testing.T, call func()) {
	t.Helper()

	done := make(chan struct{}, 1)
	go func() {
		call()
		done <- struct{}{}
	}()

	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		t.Fatal("Locale call deadlocked while translator mutated its Locale")
	}
}

func (tr *reentrantTranslator) record(name string) {
	tr.calls[name]++
}

func (tr *reentrantTranslator) mutateLocale() {
	if tr.locale != nil {
		tr.locale.SetDomain("reentrant")
	}
}

func (tr *reentrantTranslator) ParseFile(f string) {
	tr.record("ParseFile")
	tr.po.ParseFile(f)
}

func (tr *reentrantTranslator) Parse(data []byte) {
	tr.record("Parse")
	tr.po.Parse(data)
}

func (tr *reentrantTranslator) Get(str string, vars ...any) string {
	tr.record("Get")
	tr.mutateLocale()
	return tr.po.Get(str, vars...)
}

func (tr *reentrantTranslator) GetN(str, plural string, n int, vars ...any) string {
	tr.record("GetN")
	tr.mutateLocale()
	return tr.po.GetN(str, plural, n, vars...)
}

func (tr *reentrantTranslator) GetC(str, ctx string, vars ...any) string {
	tr.record("GetC")
	tr.mutateLocale()
	return tr.po.GetC(str, ctx, vars...)
}

func (tr *reentrantTranslator) GetNC(str, plural string, n int, ctx string, vars ...any) string {
	tr.record("GetNC")
	tr.mutateLocale()
	return tr.po.GetNC(str, plural, n, ctx, vars...)
}

func (tr *reentrantTranslator) MarshalBinary() ([]byte, error) {
	tr.record("MarshalBinary")
	tr.mutateLocale()
	if tr.marshalErr != nil {
		return nil, tr.marshalErr
	}
	return tr.po.MarshalBinary()
}
func (tr *reentrantTranslator) UnmarshalBinary(data []byte) error {
	tr.record("UnmarshalBinary")
	return tr.po.UnmarshalBinary(data)
}

func (tr *reentrantTranslator) GetDomain() *Domain {
	tr.record("GetDomain")
	tr.mutateLocale()
	return tr.po.GetDomain()
}

func TestLocale(t *testing.T) {
	// Set PO content
	str := `
msgid ""
msgstr ""
# Initial comment
# Headers below
"Language: en\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

# Some comment
msgid "My text"
msgstr "Translated text"

# More comments
msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular: %s"
msgstr[1] "This one is the plural: %s"
msgstr[2] "And this is the second plural form: %s"

msgid "This one has invalid syntax translations"
msgid_plural "Plural index"
msgstr[abc] "Wrong index"
msgstr[1 "Forgot to close brackets"
msgstr[0] "Badly formatted string'

msgctxt "Ctx"
msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular in a Ctx context: %s"
msgstr[1] "This one is the plural in a Ctx context: %s"

msgid "Some random"
msgstr "Some random Translation"

msgctxt "Ctx"
msgid "Some random in a context"
msgstr "Some random Translation in a context"

msgid "More"
msgstr "More Translation"

	`

	var locales []*Locale
	{ // test os
		// Create Locales directory with simplified language code
		dirname := path.Join("/tmp", "en", "LC_MESSAGES")
		err := os.MkdirAll(dirname, os.ModePerm)
		if err != nil {
			t.Fatalf("Can't create test directory: %s", err.Error())
		}

		// Write PO content to file
		filename := path.Join(dirname, "my_domain.po")

		f, err := os.Create(filename)
		if err != nil {
			t.Fatalf("Can't create test file: %s", err.Error())
		}
		defer func() {
			_ = f.Close()
		}()

		_, err = f.WriteString(str)
		if err != nil {
			t.Fatalf("Can't write to test file: %s", err.Error())
		}

		// Create Locale with full language code
		locales = append(locales, NewLocale("/tmp", "en_US"))
	}
	{ // test fs
		fs := make(fstest.MapFS)
		fs["en/LC_MESSAGES/my_domain.po"] = &fstest.MapFile{Data: []byte(str)}
		locales = append(locales, NewLocaleFS("en_US", fs))
	}

	for _, l := range locales {
		// Force nil domain storage
		l.Domains = nil

		// Add domain
		l.AddDomain("my_domain")

		// Test translations
		tr := l.GetD("my_domain", "My text")
		if tr != translatedText {
			t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
		}

		v := "Variable"
		tr = l.GetD("my_domain", "One with var: %s", v)
		if tr != "This one is the singular: Variable" {
			t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
		}

		// Test plural
		tr = l.GetND("my_domain", "One with var: %s", "Several with vars: %s", 7, v)
		if tr != "This one is the plural: Variable" {
			t.Errorf("Expected 'This one is the plural: Variable' but got '%s'", tr)
		}

		// Test context translations
		tr = l.GetC("Some random in a context", "Ctx")
		if tr != "Some random Translation in a context" {
			t.Errorf("Expected 'Some random Translation in a context'. Got '%s'", tr)
		}

		v = "Test"
		tr = l.GetNC("One with var: %s", "Several with vars: %s", 23, "Ctx", v)
		if tr != "This one is the plural in a Ctx context: Test" {
			t.Errorf("Expected 'This one is the plural in a Ctx context: Test'. Got '%s'", tr)
		}

		tr = l.GetDC("my_domain", "One with var: %s", "Ctx", v)
		if tr != "This one is the singular in a Ctx context: Test" {
			t.Errorf("Expected 'This one is the singular in a Ctx context: Test' but got '%s'", tr)
		}

		// Test plural
		tr = l.GetNDC("my_domain", "One with var: %s", "Several with vars: %s", 3, "Ctx", v)
		if tr != "This one is the plural in a Ctx context: Test" {
			t.Errorf("Expected 'This one is the plural in a Ctx context: Test' but got '%s'", tr)
		}

		// Test last Translation
		tr = l.GetD("my_domain", "More")
		if tr != "More Translation" {
			t.Errorf("Expected 'More Translation' but got '%s'", tr)
		}
	}
}

func TestLocaleFails(t *testing.T) {
	// Set PO content
	str := `
msgid ""
msgstr ""
# Initial comment
# Headers below
"Language: en\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

# Some comment
msgid "My text"
msgstr "Translated text"

# More comments
msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular: %s"
msgstr[1] "This one is the plural: %s"
msgstr[2] "And this is the second plural form: %s"

msgid "This one has invalid syntax translations"
msgid_plural "Plural index"
msgstr[abc] "Wrong index"
msgstr[1 "Forgot to close brackets"
msgstr[0] "Badly formatted string'

msgid "Invalid formatted id[] with no translations

msgctxt "Ctx"
msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular in a Ctx context: %s"
msgstr[1] "This one is the plural in a Ctx context: %s"

msgid "Some random"
msgstr "Some random Translation"

msgctxt "Ctx"
msgid "Some random in a context"
msgstr "Some random Translation in a context"

msgid "More"
msgstr "More Translation"

	`

	// Create Locales directory with simplified language code
	dirname := path.Join("/tmp", "en", "LC_MESSAGES")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to file
	filename := path.Join(dirname, "my_domain.po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Create Locale with full language code
	l := NewLocale("/tmp", "en_US")

	// Test language
	language := l.GetLanguage()
	if language != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", language)
	}

	// Force nil domain storage
	l.Domains = nil

	// Add domain
	l.AddDomain("my_domain")

	// Test non-existent "default" domain responses
	tr := l.GetDomain()
	if tr != "my_domain" {
		t.Errorf("Expected 'my_domain' but got '%s'", tr)
	}

	// Set default domain to make it fail
	l.SetDomain("default")

	// Test non-existent "default" domain responses
	tr = l.GetDomain()
	if tr != "default" {
		t.Errorf("Expected 'default' but got '%s'", tr)
	}

	// Test non-existent "default" domain responses
	tr = l.Get("My text")
	if tr != "My text" {
		t.Errorf("Expected 'My text' but got '%s'", tr)
	}

	v := "Variable"
	tr = l.GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "Several with vars: Variable" {
		t.Errorf("Expected 'Several with vars: Variable' but got '%s'", tr)
	}

	// Test inexistent translations
	tr = l.Get("This is a test")
	if tr != "This is a test" {
		t.Errorf("Expected 'This is a test' but got '%s'", tr)
	}

	tr = l.GetN("This is a test", "This are tests", 1)
	if tr != "This is a test" {
		t.Errorf("Expected 'This is a test' but got '%s'", tr)
	}

	tr = l.GetN("This is a test", "This are tests", 7)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Test syntax error parsed translations
	tr = l.Get("This one has invalid syntax translations")
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}

	tr = l.GetN("This one has invalid syntax translations", "This are tests", 1)
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}
	tr = l.GetN("This one has invalid syntax translations", "This are tests", 2)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Create Locale with full language code
	l = NewLocale("/tmp", "golem")

	// Test language
	language = l.GetLanguage()
	if language != "golem" {
		t.Errorf("Expected 'golem' but got '%s'", language)
	}

	// Force nil domain storage
	l.Domains = nil

	// Add domain
	l.SetDomain("my_domain")

	// Test non-existent "default" domain responses
	tr = l.GetDomain()
	if tr != "my_domain" {
		t.Errorf("Expected 'my_domain' but got '%s'", tr)
	}

	// Test syntax error parsed translations
	tr = l.Get("This one has invalid syntax translations")
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}

	tr = l.GetN("This one has invalid syntax translations", "This are tests", 1)
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}
	tr = l.GetN("This one has invalid syntax translations", "This are tests", 111)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Create Locale with full language code
	l = NewLocale("fixtures/", "fr_FR")

	// Test language
	language = l.GetLanguage()
	if language != "fr_FR" {
		t.Errorf("Expected 'fr_FR' but got '%s'", language)
	}

	// Force nil domain storage
	l.Domains = nil

	// Add domain
	l.SetDomain("default")

	// Test non-existent "default" domain responses
	tr = l.GetDomain()
	if tr != "default" {
		t.Errorf("Expected 'my_domain' but got '%s'", tr)
	}

	// Test syntax error parsed translations
	tr = l.Get("This one has invalid syntax translations")
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}

	tr = l.GetN("This one has invalid syntax translations", "This are tests", 1)
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}
	tr = l.GetN("This one has invalid syntax translations", "This are tests", 21)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Create Locale with full language code
	l = NewLocale("fixtures/", "de_DE")

	// Test language
	language = l.GetLanguage()
	if language != "de_DE" {
		t.Errorf("Expected 'de_DE' but got '%s'", language)
	}

	// Force nil domain storage
	l.Domains = nil

	// Add domain
	l.SetDomain("default")

	// Test non-existent "default" domain responses
	tr = l.GetDomain()
	if tr != "default" {
		t.Errorf("Expected 'my_domain' but got '%s'", tr)
	}

	// Test syntax error parsed translations
	tr = l.Get("This one has invalid syntax translations")
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}

	tr = l.GetN("This one has invalid syntax translations", "This are tests", 1)
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}
	tr = l.GetN("This one has invalid syntax translations", "This are tests", 2)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Create Locale with full language code
	l = NewLocale("fixtures/", "de_AT")

	// Test language
	language = l.GetLanguage()
	if language != "de_AT" {
		t.Errorf("Expected 'de_AT' but got '%s'", language)
	}

	// Force nil domain storage
	l.Domains = nil

	// Add domain
	l.SetDomain("default")

	// Test non-existent "default" domain responses
	tr = l.GetDomain()
	if tr != "default" {
		t.Errorf("Expected 'my_domain' but got '%s'", tr)
	}

	// Test syntax error parsed translations
	tr = l.Get("This one has invalid syntax translations")
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}

	// Test syntax error parsed translations
	tr = l.GetNDC("mega", "This one has invalid syntax translations", "plural", 2, "ctx")
	if tr != "plural" {
		t.Errorf("Expected 'plural' but got '%s'", tr)
	}

	tr = l.GetN("This one has invalid syntax translations", "This are tests", 1)
	if tr != "This one has invalid syntax translations" {
		t.Errorf("Expected 'This one has invalid syntax translations' but got '%s'", tr)
	}
	tr = l.GetN("This one has invalid syntax translations", "This are tests", 14)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}
}

func TestLocaleRace(t *testing.T) {
	// Set PO content
	str := `# Some comment
msgid "My text"
msgstr "Translated text"

# More comments
msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular: %s"
msgstr[1] "This one is the plural: %s"
msgstr[2] "And this is the second plural form: %s"

	`

	var locales []*Locale
	{ // test os
		// Create Locales directory with simplified language code
		dirname := path.Join("/tmp", "es")
		err := os.MkdirAll(dirname, os.ModePerm)
		if err != nil {
			t.Fatalf("Can't create test directory: %s", err.Error())
		}

		// Write PO content to file
		filename := path.Join(dirname, "race.po")

		f, err := os.Create(filename)
		if err != nil {
			t.Fatalf("Can't create test file: %s", err.Error())
		}
		defer func() {
			_ = f.Close()
		}()

		_, err = f.WriteString(str)
		if err != nil {
			t.Fatalf("Can't write to test file: %s", err.Error())
		}

		// Create Locale
		locales = append(locales, NewLocale("/tmp", "es"))
	}
	{ // test fs
		fs := make(fstest.MapFS)
		fs["es/LC_MESSAGES/race.po"] = &fstest.MapFile{Data: []byte(str)}
		locales = append(locales, NewLocaleFS("es", fs))
	}

	for _, l := range locales {
		// Init sync channels
		ac := make(chan bool)
		rc := make(chan bool)

		// Add domain in goroutine
		go func(l *Locale, done chan bool) {
			l.AddDomain("race")
			done <- true
		}(l, ac)

		// Get translations in goroutine
		go func(l *Locale, done chan bool) {
			l.GetD("race", "My text")
			done <- true
		}(l, rc)

		// Get translations at top level
		l.GetD("race", "My text")

		// Wait for goroutines to finish
		<-ac
		<-rc
	}
}

func TestAddTranslator(t *testing.T) {
	// Create po object
	po := NewPo()

	// Parse file
	po.ParseFile("fixtures/en_US/default.po")

	// Create Locale
	l := NewLocale("", "en")

	// Add PO Translator to Locale object
	l.AddTranslator("default", po)

	// Test translations
	tr := l.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = l.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestAddTranslatorFS(t *testing.T) {
	fs := os.DirFS("fixtures")
	// Create po object
	po := NewPoFS(fs)

	// Parse file
	po.ParseFile("en_US/default.po")

	// Create Locale
	l := NewLocaleFS("en", fs)

	// Add PO Translator to Locale object
	l.AddTranslator("default", po)

	// Test translations
	tr := l.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = l.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestArabicTranslation(t *testing.T) {
	var locales []*Locale
	{ // test os
		// Create Locale
		locales = append(locales, NewLocale("fixtures/", "ar"))
	}
	{ // test fs
		locales = append(locales, NewLocaleFS("ar", os.DirFS("fixtures")))
	}
	for _, l := range locales {

		// Add domain
		l.AddDomain("categories")

		// Plurals formula missing + Plural translation string missing
		tr := l.GetD("categories", "Alcohol & Tobacco")
		if tr != "الكحول والتبغ" {
			t.Errorf("Expected to get 'الكحول والتبغ', but got '%s'", tr)
		}

		// Plural translation string present without translations, should get the msgid_plural
		tr = l.GetND("categories", "%d selected", "%d selected", 10)
		if tr != "%d selected" {
			t.Errorf("Expected to get '%%d selected', but got '%s'", tr)
		}

		//Plurals formula present + Plural translation string present and complete
		tr = l.GetND("categories", "Load %d more document", "Load %d more documents", 0)
		if tr != "حمّل %d مستندات إضافيّة" {
			t.Errorf("Expected to get 'msgstr[0]', but got '%s'", tr)
		}

		tr = l.GetND("categories", "Load %d more document", "Load %d more documents", 1)
		if tr != "حمّل مستند واحد إضافي" {
			t.Errorf("Expected to get 'msgstr[1]', but got '%s'", tr)
		}

		tr = l.GetND("categories", "Load %d more document", "Load %d more documents", 2)
		if tr != "حمّل مستندين إضافيين" {
			t.Errorf("Expected to get 'msgstr[2]', but got '%s'", tr)
		}

		tr = l.GetND("categories", "Load %d more document", "Load %d more documents", 6)
		if tr != "حمّل %d مستندات إضافيّة" {
			t.Errorf("Expected to get 'msgstr[3]', but got '%s'", tr)
		}

		tr = l.GetND("categories", "Load %d more document", "Load %d more documents", 116)
		if tr != "حمّل %d مستندا إضافيّا" {
			t.Errorf("Expected to get 'msgstr[4]', but got '%s'", tr)
		}

		tr = l.GetND("categories", "Load %d more document", "Load %d more documents", 102)
		if tr != "حمّل %d مستند إضافي" {
			t.Errorf("Expected to get 'msgstr[5]', but got '%s'", tr)
		}
	}
}

func TestArabicMissingPluralForm(t *testing.T) {
	var locales []*Locale
	{ // test os
		// Create Locale
		locales = append(locales, NewLocale("fixtures/", "ar"))
	}
	{ // test fs
		locales = append(locales, NewLocaleFS("ar", os.DirFS("fixtures")))
	}

	for _, l := range locales {
		// Add domain
		l.AddDomain("no_plural_header")

		// Get translation
		tr := l.GetD("no_plural_header", "Alcohol & Tobacco")
		if tr != "الكحول والتبغ" {
			t.Errorf("Expected to get 'الكحول والتبغ', but got '%s'", tr)
		}
	}
}

func TestLocaleBinaryEncoding(t *testing.T) {
	const pluralForms = "nplurals=2; plural=(n != 1);"
	data := encodeTestGob(t, &LocaleEncoding{
		Path:          "fixtures/",
		Lang:          "en_US",
		DefaultDomain: "default",
		Domains: map[string][]byte{
			"default": encodeTestGob(t, &TranslatorEncoding{
				Language:    "en_US",
				PluralForms: pluralForms,
				Nplurals:    2,
				Plural:      "(n != 1)",
				Translations: map[string]*Translation{
					"My text": {
						ID:  "My text",
						Trs: map[int]string{0: "Translated text"},
					},
					"More": {
						ID:  "More",
						Trs: map[int]string{0: "More translation"},
					},
					"One with var: %s": {
						ID:       "One with var: %s",
						PluralID: "Several with vars: %s",
						Trs:      map[int]string{0: "This one is the singular: %s", 1: "This one is the plural: %s"},
					},
				},
			}),
		},
	})

	locale := new(Locale)
	if err := locale.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if got := locale.GetLanguage(); got != "en_US" {
		t.Errorf("decoded language = %q, want %q", got, "en_US")
	}
	if got := locale.GetDomain(); got != "default" {
		t.Errorf("decoded default domain = %q, want %q", got, "default")
	}
	if got := locale.Get("My text"); got != "Translated text" {
		t.Errorf("decoded My text = %q, want %q", got, "Translated text")
	}
	if got := locale.Get("More"); got != "More translation" {
		t.Errorf("decoded More = %q, want %q", got, "More translation")
	}
	if got := locale.GetN("One with var: %s", "Several with vars: %s", 3, "VALUE"); got != "This one is the plural: VALUE" {
		t.Errorf("decoded plural = %q, want %q", got, "This one is the plural: VALUE")
	}
}

//go:embed fixtures
var localeFS embed.FS

func TestLocale_GetTranslations(t *testing.T) {
	var locales []*Locale
	{ // test os
		locales = append(locales, NewLocale("fixtures/", "en_US"))
	}
	{ // test fs
		locales = append(locales, NewLocaleFS("en_US", os.DirFS("fixtures")))
	}
	{ // test embed
		locales = append(locales, NewLocaleFSWithPath("en_US", localeFS, "fixtures"))
	}

	for _, l := range locales {
		l.AddDomain("default")

		all := l.GetTranslations()

		if len(all) < 5 {
			t.Errorf("length of all translations is too few: %d", len(all))
		}

		const moreMsgID = "More"
		more, ok := all[moreMsgID]
		if !ok {
			t.Error("missing expected translation")
		}
		if got := more.Get(); got != "More translation" {
			t.Errorf("translations of msgid %s = %q, want %q", moreMsgID, got, "More translation")
		}
		more.Set("caller mutation")
		if got := l.Get(moreMsgID); got != "More translation" {
			t.Errorf("mutating GetTranslations result changed Locale: got %q", got)
		}
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
func TestLocaleTranslatorCallsReleaseLock(t *testing.T) {
	locale := NewLocale("", "en")
	translator := newReentrantTranslator(locale)
	locale.AddTranslator("default", translator)

	calls := []struct {
		name           string
		translatorCall string
		call           func()
	}{
		{name: "Get", translatorCall: "Get", call: func() { _ = locale.Get("message") }},
		{name: "GetN", translatorCall: "GetN", call: func() { _ = locale.GetN("message", "messages", 2) }},
		{name: "GetC", translatorCall: "GetC", call: func() { _ = locale.GetC("message", "context") }},
		{name: "GetNC", translatorCall: "GetNC", call: func() { _ = locale.GetNC("message", "messages", 2, "context") }},
		{name: "GetD", translatorCall: "Get", call: func() { _ = locale.GetD("default", "message") }},
		{name: "GetND", translatorCall: "GetN", call: func() { _ = locale.GetND("default", "message", "messages", 2) }},
		{name: "GetDC", translatorCall: "GetC", call: func() { _ = locale.GetDC("default", "message", "context") }},
		{name: "GetNDC", translatorCall: "GetNC", call: func() { _ = locale.GetNDC("default", "message", "messages", 2, "context") }},
		{name: "IsTranslatedND", translatorCall: "GetDomain", call: func() { _ = locale.IsTranslatedND("default", "message", 0) }},
		{name: "IsTranslatedNDC", translatorCall: "GetDomain", call: func() { _ = locale.IsTranslatedNDC("default", "message", 0, "context") }},
		{name: "GetTranslations", translatorCall: "GetDomain", call: func() { _ = locale.GetTranslations() }},
		{name: "MarshalBinary", translatorCall: "MarshalBinary", call: func() { _, _ = locale.MarshalBinary() }},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			locale.SetDomain("default")
			before := translator.calls[tt.translatorCall]
			waitForLocaleCall(t, tt.call)
			if got := translator.calls[tt.translatorCall]; got != before+1 {
				t.Fatalf("%s did not invoke translator.%s exactly once: count %d -> %d", tt.name, tt.translatorCall, before, got)
			}
		})
	}

	if got := locale.GetDomain(); got != "reentrant" {
		t.Errorf("reentrant translator did not mutate Locale: got domain %q", got)
	}
}

func TestLocaleMarshalBinaryPropagatesTranslatorError(t *testing.T) {
	sentinel := errors.New("sentinel marshal error")
	locale := NewLocale("", "en")
	translator := newReentrantTranslator(locale)
	translator.marshalErr = sentinel
	locale.AddTranslator("default", translator)

	var gotErr error
	waitForLocaleCall(t, func() {
		_, gotErr = locale.MarshalBinary()
	})
	if !errors.Is(gotErr, sentinel) {
		t.Fatalf("MarshalBinary error = %v, want sentinel %v", gotErr, sentinel)
	}
}
func TestLocaleNilTranslatorIsSafe(t *testing.T) {
	locale := NewLocale("", "en")
	locale.AddTranslator("nil", nil)

	if got := locale.GetD("nil", "source"); got != "source" {
		t.Errorf("GetD with nil translator = %q, want %q", got, "source")
	}
	if got := locale.GetND("nil", "one", "many", 2); got != "many" {
		t.Errorf("GetND with nil translator = %q, want %q", got, "many")
	}
	if got := locale.GetDC("nil", "source", "context"); got != "source" {
		t.Errorf("GetDC with nil translator = %q, want %q", got, "source")
	}
	if got := locale.GetNDC("nil", "one", "many", 2, "context"); got != "many" {
		t.Errorf("GetNDC with nil translator = %q, want %q", got, "many")
	}
	if locale.IsTranslatedND("nil", "source", 0) {
		t.Error("IsTranslatedND with nil translator should be false")
	}
	if locale.IsTranslatedNDC("nil", "source", 0, "context") {
		t.Error("IsTranslatedNDC with nil translator should be false")
	}
	if got := locale.GetTranslations(); got == nil {
		t.Error("GetTranslations with nil translator returned nil")
	}
	if _, err := locale.MarshalBinary(); err != nil {
		t.Fatalf("MarshalBinary with nil translator failed: %v", err)
	}
}

func TestLocaleBinaryEncodingRestoresPluralForms(t *testing.T) {
	const (
		msgID      = "One apple"
		msgPlural  = "Many apples"
		pluralRule = "(n==1) ? 0 : (n==2) ? 1 : 2"
	)

	data := encodeTestGob(t, &LocaleEncoding{
		Lang:          "xx",
		DefaultDomain: "default",
		Domains: map[string][]byte{
			"default": encodeTestGob(t, &TranslatorEncoding{
				Language:    "xx",
				PluralForms: "nplurals=3; plural=" + pluralRule + ";",
				Nplurals:    3,
				Plural:      pluralRule,
				Translations: map[string]*Translation{
					msgID: {
						ID:       msgID,
						PluralID: msgPlural,
						Trs:      map[int]string{0: "one apple", 1: "two apples", 2: "many apples"},
					},
				},
			}),
		},
	})

	restored := new(Locale)
	if err := restored.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if got := restored.GetLanguage(); got != "xx" {
		t.Errorf("decoded language = %q, want %q", got, "xx")
	}
	if got := restored.GetN(msgID, msgPlural, 1); got != "one apple" {
		t.Errorf("GetN(1) = %q, want %q", got, "one apple")
	}
	if got := restored.GetN(msgID, msgPlural, 2); got != "two apples" {
		t.Errorf("GetN(2) = %q, want %q", got, "two apples")
	}
	if got := restored.GetN(msgID, msgPlural, 5); got != "many apples" {
		t.Errorf("GetN(5) = %q, want %q", got, "many apples")
	}
}

func TestLocaleBinaryEncodingInvalidPluralFallback(t *testing.T) {
	data := encodeTestGob(t, &LocaleEncoding{
		Lang:          "xx",
		DefaultDomain: "default",
		Domains: map[string][]byte{
			"default": encodeTestGob(t, &TranslatorEncoding{
				Language:    "xx",
				PluralForms: "nplurals=3; plural=invalid;",
				Nplurals:    3,
				Plural:      "invalid",
				Translations: map[string]*Translation{
					"One apple": {
						ID:       "One apple",
						PluralID: "Many apples",
						Trs:      map[int]string{0: "one apple", 1: "many apples"},
					},
				},
			}),
		},
	})

	restored := new(Locale)
	if err := restored.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if got := restored.GetN("One apple", "Many apples", 1); got != "one apple" {
		t.Errorf("invalid plural rule GetN(1) = %q, want %q", got, "one apple")
	}
	if got := restored.GetN("One apple", "Many apples", 2); got != "many apples" {
		t.Errorf("invalid plural rule GetN(2) = %q, want %q", got, "many apples")
	}
}

func TestTranslatorEncodingGetTranslatorInitializesStorage(t *testing.T) {
	initialTranslation := NewTranslation()
	initialTranslation.ID = "initial"
	initialTranslation.Set("initial translation")

	initialContextTranslation := NewTranslation()
	initialContextTranslation.ID = "initial context"
	initialContextTranslation.Set("initial context translation")

	partial := TranslatorEncoding{
		Headers: HeaderMap{
			"X-Initial": {"initial header"},
		},
		Translations: map[string]*Translation{
			"initial": initialTranslation,
		},
		Contexts: map[string]map[string]*Translation{
			"initial-context": {
				"initial context": initialContextTranslation,
			},
		},
	}

	tests := []struct {
		name     string
		encoding TranslatorEncoding
	}{
		{name: "zero", encoding: TranslatorEncoding{}},
		{name: "partial", encoding: partial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po, ok := tt.encoding.GetTranslator().(*Po)
			if !ok {
				t.Fatalf("GetTranslator() returned %T, want *Po", tt.encoding.GetTranslator())
			}

			po.Set("set message", "set translation")
			if got := po.Get("set message"); got != "set translation" {
				t.Errorf("Get after Set = %q, want %q", got, "set translation")
			}

			po.SetC("set context message", "set-context", "set context translation")
			if got := po.GetC("set context message", "set-context"); got != "set context translation" {
				t.Errorf("GetC after SetC = %q, want %q", got, "set context translation")
			}

			po.Parse([]byte(`
msgid ""
msgstr ""
"Language: xx\n"
"X-Test: header value\n"

msgid "parsed message"
msgstr "parsed translation"
`))

			if got := po.Get("parsed message"); got != "parsed translation" {
				t.Errorf("Get after Parse = %q, want %q", got, "parsed translation")
			}
			if got := po.Get("set message"); got != "set translation" {
				t.Errorf("Get lost value set before Parse: %q", got)
			}
			if got := po.GetC("set context message", "set-context"); got != "set context translation" {
				t.Errorf("GetC lost value set before Parse: %q", got)
			}
			if got := po.Headers.Get("Language"); got != "xx" {
				t.Errorf("public header access = %q, want %q", got, "xx")
			}
			if got := po.GetDomain().Headers.Get("X-Test"); got != "header value" {
				t.Errorf("domain header access = %q, want %q", got, "header value")
			}
			if po.Language != "xx" {
				t.Errorf("public Language = %q, want %q", po.Language, "xx")
			}

			if tt.name != "partial" {
				return
			}

			transferredTranslation := NewTranslation()
			transferredTranslation.ID = "transferred"
			transferredTranslation.Set("transferred translation")
			tt.encoding.Headers["X-Transferred"] = []string{"transferred header"}
			tt.encoding.Translations["transferred"] = transferredTranslation
			tt.encoding.Contexts["transferred-context"] = map[string]*Translation{
				"transferred context": transferredTranslation,
			}

			if got := po.Headers.Get("X-Transferred"); got != "transferred header" {
				t.Errorf("public Headers did not retain map ownership: %q", got)
			}
			if got := po.Get("transferred"); got != "transferred translation" {
				t.Errorf("Translations did not retain map ownership: %q", got)
			}
			if got := po.GetC("transferred context", "transferred-context"); got != "transferred translation" {
				t.Errorf("Contexts did not retain map ownership: %q", got)
			}
		})
	}
}
func TestTranslatorEncodingNormalizesNilEntries(t *testing.T) {
	ordinary := map[string]*Translation{
		"ordinary": nil,
	}
	contexts := map[string]map[string]*Translation{
		"nil-context": nil,
		"context": {
			"contextual": nil,
		},
	}
	encoding := &TranslatorEncoding{
		Translations: ordinary,
		Contexts:     contexts,
	}

	po, ok := encoding.GetTranslator().(*Po)
	if !ok {
		t.Fatalf("GetTranslator() returned %T, want *Po", encoding.GetTranslator())
	}

	if ordinary["ordinary"] == nil {
		t.Fatal("ordinary nil entry was not normalized in the caller-owned map")
	}
	if got := po.Get("ordinary"); got != "ordinary" {
		t.Errorf("normalized ordinary entry = %q, want %q", got, "ordinary")
	}
	if contexts["nil-context"] == nil {
		t.Fatal("nil context map was not normalized in the caller-owned map")
	}
	if contexts["context"]["contextual"] == nil {
		t.Fatal("nil contextual entry was not normalized in the caller-owned map")
	}
	if got := po.GetC("contextual", "context"); got != "contextual" {
		t.Errorf("normalized contextual entry = %q, want %q", got, "contextual")
	}

	transferred := NewTranslation()
	transferred.ID = "transferred"
	transferred.Set("transferred value")
	ordinary["transferred"] = transferred
	if got := po.Get("transferred"); got != "transferred value" {
		t.Errorf("non-nil translation map was not transferred by ownership: got %q", got)
	}
}

func TestLocaleBinaryEncodingKeepsStateOnDomainDecodeError(t *testing.T) {
	filesystem := fstest.MapFS{}
	locale := NewLocaleFSWithPath("old", filesystem, "old/path")
	oldTranslator := NewPo()
	locale.AddTranslator("old", oldTranslator)
	locale.SetDomain("old")

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(&LocaleEncoding{
		Path:          "new/path",
		Lang:          "new",
		Domains:       map[string][]byte{"new": []byte("invalid")},
		DefaultDomain: "new",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := locale.UnmarshalBinary(buff.Bytes()); err == nil {
		t.Fatal("UnmarshalBinary returned nil for an invalid domain")
	}

	if locale.path != "old/path" {
		t.Errorf("path changed after failed decode: %q", locale.path)
	}
	if locale.GetLanguage() != "old" {
		t.Errorf("language changed after failed decode: %q", locale.GetLanguage())
	}
	if locale.GetDomain() != "old" {
		t.Errorf("default domain changed after failed decode: %q", locale.GetDomain())
	}
	if locale.fs == nil {
		t.Error("filesystem was cleared after failed decode")
	}
	if got, ok := locale.Domains["old"]; !ok || got != oldTranslator {
		t.Error("existing domains changed after failed decode")
	}
	if _, ok := locale.Domains["new"]; ok {
		t.Error("partially decoded domain was installed after failed decode")
	}
}

func TestLocaleBinaryEncodingRace(t *testing.T) {
	const iterations = 100

	source := NewLocale("", "en")
	source.AddTranslator("default", NewPo())
	target := NewLocale("", "en")

	var wg sync.WaitGroup
	wg.Add(3)

	// Each replacement uses a fresh translator so concurrent operations do not
	// share mutable translator state.
	go func() {
		defer wg.Done()
		for i := range iterations {
			domain := "domain-a"
			if i%2 == 0 {
				domain = "domain-b"
			}
			source.AddTranslator(domain, NewPo())
			source.SetDomain(domain)
		}
	}()

	go func() {
		defer wg.Done()
		for range iterations {
			data, err := source.MarshalBinary()
			if err != nil {
				t.Errorf("MarshalBinary() failed: %v", err)
				continue
			}
			if err := target.UnmarshalBinary(data); err != nil {
				t.Errorf("UnmarshalBinary() failed: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for range iterations {
			_ = source.GetDomain()
			_ = source.GetD("domain-a", "message")
			_ = target.GetDomain()
			_ = target.GetD("domain-b", "message")
		}
	}()

	wg.Wait()
}
func FuzzDomainUnmarshalBinaryAtomicAndUsable(f *testing.F) {
	valid := encodeTestGob(f, &TranslatorEncoding{
		Translations: map[string]*Translation{
			"seed": {
				ID:  "seed",
				Trs: map[int]string{0: "value"},
			},
		},
	})
	f.Add(valid)
	f.Add([]byte("not gob"))
	f.Add([]byte{})
	f.Add([]byte{0, 1, 2, 3})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}

		domain := NewDomain()
		domain.Set("sentinel", "before")

		var unmarshalErr error
		panicked := false
		func() {
			defer func() {
				if recover() != nil {
					panicked = true
				}
			}()
			unmarshalErr = domain.UnmarshalBinary(data)
		}()
		if panicked {
			t.Fatal("UnmarshalBinary panicked")
		}

		if unmarshalErr != nil {
			if got := domain.Get("sentinel"); got != "before" {
				t.Fatalf("failed decode changed existing state: got %q, want %q", got, "before")
			}
			return
		}

		if domain.GetTranslations() == nil {
			t.Fatal("successful decode left ordinary translations unusable")
		}
		if domain.GetCtxTranslations() == nil {
			t.Fatal("successful decode left contextual translations unusable")
		}
		_ = domain.Get("sentinel")
		_ = domain.GetN("sentinel", "sentinels", 2)
		_ = domain.GetC("sentinel", "context")
		_ = domain.GetNC("sentinel", "sentinels", 2, "context")
	})
}

func FuzzLocaleUnmarshalBinaryAtomicAndUsable(f *testing.F) {
	valid := encodeTestGob(f, &LocaleEncoding{
		Path:          "seed/path",
		Lang:          "en",
		DefaultDomain: "default",
		Domains: map[string][]byte{
			"default": encodeTestGob(f, &TranslatorEncoding{
				Translations: map[string]*Translation{
					"seed": {
						ID:  "seed",
						Trs: map[int]string{0: "value"},
					},
				},
			}),
		},
	})
	f.Add(valid)
	f.Add([]byte("not gob"))
	f.Add([]byte{})
	f.Add([]byte{0, 1, 2, 3})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}

		locale := NewLocaleFSWithPath("old", fstest.MapFS{}, "old/path")
		oldTranslator := NewPo()
		oldTranslator.Set("sentinel", "before")
		locale.AddTranslator("old", oldTranslator)
		locale.SetDomain("old")

		var unmarshalErr error
		panicked := false
		func() {
			defer func() {
				if recover() != nil {
					panicked = true
				}
			}()
			unmarshalErr = locale.UnmarshalBinary(data)
		}()
		if panicked {
			t.Fatal("UnmarshalBinary panicked")
		}

		if unmarshalErr != nil {
			if locale.path != "old/path" || locale.GetLanguage() != "old" || locale.GetDomain() != "old" {
				t.Fatal("failed decode changed locale configuration")
			}
			if locale.fs == nil {
				t.Fatal("failed decode cleared the filesystem")
			}
			if got, ok := locale.Domains["old"]; !ok || got != oldTranslator {
				t.Fatal("failed decode changed existing domains")
			}
			return
		}

		if locale.GetTranslations() == nil {
			t.Fatal("successful decode left locale translations unusable")
		}
		_ = locale.GetDomain()
		_ = locale.GetLanguage()
		_ = locale.Get("sentinel")
		_ = locale.GetN("sentinel", "sentinels", 2)
		_ = locale.GetC("sentinel", "context")
		_ = locale.GetNC("sentinel", "sentinels", 2, "context")
	})
}
