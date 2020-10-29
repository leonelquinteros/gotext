/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"os"
	"path"
	"testing"
)

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
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Create Locale with full language code
	l := NewLocale("/tmp", "en_US")

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
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Create Locale with full language code
	l := NewLocale("/tmp", "en_US")

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
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Create Locale
	l := NewLocale("/tmp", "es")

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

func TestArabicTranslation(t *testing.T) {
	// Create Locale
	l := NewLocale("fixtures/", "ar")

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

func TestArabicMissingPluralForm(t *testing.T) {
	// Create Locale
	l := NewLocale("fixtures/", "ar")

	// Add domain
	l.AddDomain("no_plural_header")

	// Get translation
	tr := l.GetD("no_plural_header", "Alcohol & Tobacco")
	if tr != "الكحول والتبغ" {
		t.Errorf("Expected to get 'الكحول والتبغ', but got '%s'", tr)
	}
}

func TestLocaleBinaryEncoding(t *testing.T) {
	// Create Locale
	l := NewLocale("fixtures/", "en_US")
	l.AddDomain("default")

	buff, err := l.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	l2 := new(Locale)
	err = l2.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	// Check object properties
	if l.path != l2.path {
		t.Fatalf("path doesn't match: '%s' vs '%s'", l.path, l2.path)
	}
	if l.lang != l2.lang {
		t.Fatalf("lang doesn't match: '%s' vs '%s'", l.lang, l2.lang)
	}
	if l.defaultDomain != l2.defaultDomain {
		t.Fatalf("defaultDomain doesn't match: '%s' vs '%s'", l.defaultDomain, l2.defaultDomain)
	}

	// Check translations
	if l.Get("My text") != l2.Get("My text") {
		t.Errorf("'%s' is different from '%s", l.Get("My text"), l2.Get("My text"))
	}
	if l.Get("More") != l2.Get("More") {
		t.Errorf("'%s' is different from '%s", l.Get("More"), l2.Get("More"))
	}
	if l.GetN("One with var: %s", "Several with vars: %s", 3, "VALUE") != l2.GetN("One with var: %s", "Several with vars: %s", 3, "VALUE") {
		t.Errorf("'%s' is different from '%s", l.GetN("One with var: %s", "Several with vars: %s", 3, "VALUE"), l2.GetN("One with var: %s", "Several with vars: %s", 3, "VALUE"))
	}
}

func TestLocale_GetTranslations(t *testing.T) {
	l := NewLocale("fixtures/", "en_US")
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
	if more.Get() != l.Get(moreMsgID) {
		t.Errorf("translations of msgid %s do not match: \"%s\" != \"%s\"", moreMsgID, more.Get(), l.Get(moreMsgID))
	}
}
