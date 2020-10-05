/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"fmt"
	"os"
	"path"
	"testing"
)

const (
	translatedText = "Translated text"
)

func TestPo_Get(t *testing.T) {
	// Create po object
	po := NewPo()

	// Try to parse a directory
	po.ParseFile(path.Clean(os.TempDir()))

	// Parse file
	po.ParseFile("fixtures/en_US/default.po")

	// Test translations
	tr := po.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = po.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestPo(t *testing.T) {
	// Set PO content
	str := `
msgid   ""
msgstr  ""

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

# Multi-line msgid
msgid ""
"multi"
"line"
"id"
msgstr "id with multiline content"

# Multi-line msgid_plural
msgid "" 
"multi"
"line"
"plural"
"id"
msgstr "plural id with multiline content"

#Multi-line string
msgid "Multi-line"
msgstr "" 
"Multi "
"line"

msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular: %s"
msgstr[1] "This one is the plural: %s"
msgstr[2] "And this is the second plural form: %s"

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

msgid "Empty Translation"
msgstr ""

msgid "Empty plural form singular"
msgid_plural "Empty plural form"
msgstr[0] "Singular translated"
msgstr[1] ""

msgid "More"
msgstr "More Translation"

	`

	// Write PO content to file
	filename := path.Clean(os.TempDir() + string(os.PathSeparator) + "default.po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Create po object
	po := NewPo()

	// Try to parse a directory
	po.ParseFile(path.Clean(os.TempDir()))

	// Parse file
	po.ParseFile(filename)

	// Test translations
	tr := po.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}

	v := "Variable"
	tr = po.Get("One with var: %s", v)
	if tr != "This one is the singular: Variable" {
		t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
	}

	// Test multi-line id
	tr = po.Get("multilineid")
	if tr != "id with multiline content" {
		t.Errorf("Expected 'id with multiline content' but got '%s'", tr)
	}

	// Test multi-line plural id
	tr = po.Get("multilinepluralid")
	if tr != "plural id with multiline content" {
		t.Errorf("Expected 'plural id with multiline content' but got '%s'", tr)
	}

	// Test multi-line
	tr = po.Get("Multi-line")
	if tr != "Multi line" {
		t.Errorf("Expected 'Multi line' but got '%s'", tr)
	}

	// Test plural
	tr = po.GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "This one is the plural: Variable" {
		t.Errorf("Expected 'This one is the plural: Variable' but got '%s'", tr)
	}

	// Test not existent translations
	tr = po.Get("This is a test")
	if tr != "This is a test" {
		t.Errorf("Expected 'This is a test' but got '%s'", tr)
	}

	tr = po.GetN("This is a test", "This are tests", 100)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Test context translations
	v = "Test"
	tr = po.GetC("One with var: %s", "Ctx", v)
	if tr != "This one is the singular in a Ctx context: Test" {
		t.Errorf("Expected 'This one is the singular in a Ctx context: Test' but got '%s'", tr)
	}

	// Test plural
	tr = po.GetNC("One with var: %s", "Several with vars: %s", 17, "Ctx", v)
	if tr != "This one is the plural in a Ctx context: Test" {
		t.Errorf("Expected 'This one is the plural in a Ctx context: Test' but got '%s'", tr)
	}

	// Test default plural vs singular return responses
	tr = po.GetN("Original", "Original plural", 4)
	if tr != "Original plural" {
		t.Errorf("Expected 'Original plural' but got '%s'", tr)
	}
	tr = po.GetN("Original", "Original plural", 1)
	if tr != "Original" {
		t.Errorf("Expected 'Original' but got '%s'", tr)
	}

	// Test empty Translation strings
	tr = po.Get("Empty Translation")
	if tr != "Empty Translation" {
		t.Errorf("Expected 'Empty Translation' but got '%s'", tr)
	}

	tr = po.Get("Empty plural form singular")
	if tr != "Singular translated" {
		t.Errorf("Expected 'Singular translated' but got '%s'", tr)
	}

	tr = po.GetN("Empty plural form singular", "Empty plural form", 1)
	if tr != "Singular translated" {
		t.Errorf("Expected 'Singular translated' but got '%s'", tr)
	}

	tr = po.GetN("Empty plural form singular", "Empty plural form", 2)
	if tr != "Empty plural form" {
		t.Errorf("Expected 'Empty plural form' but got '%s'", tr)
	}

	// Test last Translation
	tr = po.Get("More")
	if tr != "More Translation" {
		t.Errorf("Expected 'More Translation' but got '%s'", tr)
	}
}

func TestPlural(t *testing.T) {
	// Set PO content
	str := `
msgid   ""
msgstr  ""
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Singular: %s"
msgid_plural "Plural: %s"
msgstr[0] "TR Singular: %s"
msgstr[1] "TR Plural: %s"
msgstr[2] "TR Plural 2: %s"

	
`
	// Create po object
	po := NewPo()
	po.Parse([]byte(str))

	v := "Var"
	tr := po.GetN("Singular: %s", "Plural: %s", 2, v)
	if tr != "TR Plural: Var" {
		t.Errorf("Expected 'TR Plural: Var' but got '%s'", tr)
	}

	tr = po.GetN("Singular: %s", "Plural: %s", 1, v)
	if tr != "TR Singular: Var" {
		t.Errorf("Expected 'TR Singular: Var' but got '%s'", tr)
	}
}

func TestPluralNoHeaderInformation(t *testing.T) {
	// Set PO content
	str := `
msgid   ""
msgstr  ""

msgid "Singular: %s"
msgid_plural "Plural: %s"
msgstr[0] "TR Singular: %s"
msgstr[1] "TR Plural: %s"
msgstr[2] "TR Plural 2: %s"

	
`
	// Create po object
	po := NewPo()
	po.Parse([]byte(str))

	v := "Var"
	tr := po.GetN("Singular: %s", "Plural: %s", 2, v)
	if tr != "TR Plural: Var" {
		t.Errorf("Expected 'TR Plural: Var' but got '%s'", tr)
	}

	tr = po.GetN("Singular: %s", "Plural: %s", 1, v)
	if tr != "TR Singular: Var" {
		t.Errorf("Expected 'TR Singular: Var' but got '%s'", tr)
	}
}

func TestPoHeaders(t *testing.T) {
	// Set PO content
	str := `
msgid   ""
msgstr  ""
# Initial comment
# Headers below
"Language: en\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

# Some comment
msgid "Example"
msgstr "Translated example"
	`

	// Create po object
	po := NewPo()

	// Parse
	po.Parse([]byte(str))

	// Check headers expected
	if po.Language != "en" {
		t.Errorf("Expected 'Language: en' but got '%s'", po.Language)
	}

	do := po.GetDomain()
	// Check headers expected
	if do.PluralForms != "nplurals=2; plural=(n != 1);" {
		t.Errorf("Expected 'Plural-Forms: nplurals=2; plural=(n != 1);' but got '%s'", do.PluralForms)
	}
}

func TestMissingPoHeadersSupport(t *testing.T) {
	// Set PO content
	str := `
msgid "Example"
msgstr "Translated example"
	`

	// Create po object
	po := NewPo()

	// Parse
	po.Parse([]byte(str))

	// Check Translation expected
	if po.Get("Example") != "Translated example" {
		t.Errorf("Expected 'Translated example' but got '%s'", po.Get("Example"))
	}
}

type pluralTest struct {
	form, num int
}

func pluralExpected(t *testing.T, pluralTests []pluralTest, domain *Domain) {
	t.Helper()
	for _, pt := range pluralTests {
		pt := pt
		t.Run(fmt.Sprintf("pluralForm(%d)", pt.num), func(t *testing.T) {
			n := domain.pluralForm(pt.num)
			if n != pt.form {
				t.Errorf("Expected %d for pluralForm(%d), got %d", pt.form, pt.num, n)
			}
		})
	}
}

func TestPluralFormsSingle(t *testing.T) {
	// Single form
	str := `
msgid   ""
msgstr  ""
"Plural-Forms: nplurals=1; plural=0;"

# Some comment
msgid "Singular"
msgid_plural "Plural"
msgstr[0] "Singular form"
msgstr[1] "Plural form 1"
msgstr[2] "Plural form 2"
msgstr[3] "Plural form 3"
	`

	// Create po object
	po := NewPo()

	// Parse
	po.Parse([]byte(str))

	pluralTests := []pluralTest{
		{form: 0, num: 0},
		{form: 0, num: 1},
		{form: 0, num: 2},
		{form: 0, num: 3},
		{form: 0, num: 50},
	}

	pluralExpected(t, pluralTests, po.GetDomain())
}

func TestPluralForms2(t *testing.T) {
	// 2 forms
	str := `
msgid   ""
msgstr  ""
"Plural-Forms: nplurals=2; plural=n != 1;"

# Some comment
msgid "Singular"
msgid_plural "Plural"
msgstr[0] "Singular form"
msgstr[1] "Plural form 1"
msgstr[2] "Plural form 2"
msgstr[3] "Plural form 3"
	`

	// Create po object
	po := NewPo()

	// Parse
	po.Parse([]byte(str))

	pluralTests := []pluralTest{
		{form: 1, num: 0},
		{form: 0, num: 1},
		{form: 1, num: 2},
		{form: 1, num: 3},
	}

	pluralExpected(t, pluralTests, po.GetDomain())
}

func TestPluralForms3(t *testing.T) {
	// 3 forms
	str := `
msgid   ""
msgstr  ""
"Plural-Forms: nplurals=3; plural=n%10==1 && n%100!=11 ? 0 : n != 0 ? 1 : 2;"

# Some comment
msgid "Singular"
msgid_plural "Plural"
msgstr[0] "Singular form"
msgstr[1] "Plural form 1"
msgstr[2] "Plural form 2"
msgstr[3] "Plural form 3"
	`

	// Create po object
	po := NewPo()

	// Parse
	po.Parse([]byte(str))

	pluralTests := []pluralTest{
		{form: 2, num: 0},
		{form: 0, num: 1},
		{form: 1, num: 2},
		{form: 1, num: 3},
		{form: 1, num: 100},
		{form: 1, num: 49},
	}

	pluralExpected(t, pluralTests, po.GetDomain())
}

func TestPluralFormsSpecial(t *testing.T) {
	// 3 forms special
	str := `
msgid   ""
msgstr  ""
"Plural-Forms: nplurals=3;"
"plural=(n==1) ? 0 : (n>=2 && n<=4) ? 1 : 2;"

# Some comment
msgid "Singular"
msgid_plural "Plural"
msgstr[0] "Singular form"
msgstr[1] "Plural form 1"
msgstr[2] "Plural form 2"
msgstr[3] "Plural form 3"
	`

	// Create po object
	po := NewPo()

	// Parse
	po.Parse([]byte(str))

	pluralTests := []pluralTest{
		{form: 0, num: 1},
		{form: 1, num: 2},
		{form: 1, num: 4},
		{form: 2, num: 0},
		{form: 2, num: 1000},
	}

	pluralExpected(t, pluralTests, po.GetDomain())
}

func TestTranslationObject(t *testing.T) {
	tr := NewTranslation()
	str := tr.Get()

	if str != "" {
		t.Errorf("Expected '' but got '%s'", str)
	}

	// Set id
	tr.ID = "Text"
	str = tr.Get()

	// Get again
	if str != "Text" {
		t.Errorf("Expected 'Text' but got '%s'", str)
	}
}

func TestPoRace(t *testing.T) {
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

	// Create Po object
	po := NewPo()

	// Create sync channels
	pc := make(chan bool)
	rc := make(chan bool)

	// Parse po content in a goroutine
	go func(po *Po, done chan bool) {
		po.Parse([]byte(str))
		done <- true
	}(po, pc)

	// Read some Translation on a goroutine
	go func(po *Po, done chan bool) {
		po.Get("My text")
		done <- true
	}(po, rc)

	// Read something at top level
	po.Get("My text")

	// Wait for goroutines to finish
	<-pc
	<-rc
}

func TestNewPoTranslatorRace(t *testing.T) {
	// Create Po object
	po := NewPo()

	// Create sync channels
	pc := make(chan bool)
	rc := make(chan bool)

	// Parse po content in a goroutine
	go func(mo Translator, done chan bool) {
		// Parse file
		mo.ParseFile("fixtures/en_US/default.po")
		done <- true
	}(po, pc)

	// Read some Translation on a goroutine
	go func(mo Translator, done chan bool) {
		mo.Get("My text")
		done <- true
	}(po, rc)

	// Read something at top level
	po.Get("My text")

	// Wait for goroutines to finish
	<-pc
	<-rc
}

func TestPoBinaryEncoding(t *testing.T) {
	// Create po objects
	po := NewPo()
	po2 := NewPo()

	// Parse file
	po.ParseFile("fixtures/en_US/default.po")

	buff, err := po.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = po2.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	// Test translations
	tr := po2.Get("My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}
	// Test translations
	tr = po2.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestPoTextEncoding(t *testing.T) {
	// Create po objects
	po := NewPo()
	po2 := NewPo()

	// Parse file
	po.ParseFile("fixtures/en_US/default.po")

	if _, ok := po.Headers["Pot-Creation-Date"]; ok {
		t.Errorf("Expected non-canonicalised header, got canonicalised")
	} else {
		if _, ok = po.Headers["POT-Creation-Date"]; !ok {
			t.Errorf("Expected non-canonicalised header, but it was missing")
		}
	}

	// Round-trip
	buff, err := po.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	po2.Parse(buff)

	for k, v := range po.Headers {
		if v2, ok := po2.Headers[k]; ok {
			for i, value := range v {
				if value != v2[i] {
					t.Errorf("TestPoTextEncoding: Header Difference for %s: %s vs %s", k, value, v2[i])
				}
			}
		}
	}

	// Test translations
	tr := po2.Get("My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}

	tr = po2.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}

	tr = po2.Get("Some random")
	if tr != "Some random translation" {
		t.Errorf("Expected 'Some random translation' but got '%s'", tr)
	}

	v := "Test"
	tr = po.GetC("One with var: %s", "Ctx", v)
	if tr != "This one is the singular in a Ctx context: Test" {
		t.Errorf("Expected 'This one is the singular in a Ctx context: Test' but got '%s'", tr)
	}

	tr = po.GetNC("One with var: %s", "Several with vars: %s", 17, "Ctx", v)
	if tr != "This one is the plural in a Ctx context: Test" {
		t.Errorf("Expected 'This one is the plural in a Ctx context: Test' but got '%s'", tr)
	}

	// Another kind of round-trip
	po.Set("My text", "Translated text")
	po.Set("language", "en_US")

	// But remove 'the'
	po.SetNC("One with var: %s", "Several with vars: %s", "Ctx", 1, "This one is singular in a Ctx context: %s")
	po.SetNC("One with var: %s", "Several with vars: %s", "Ctx", 17, "This one is plural in a Ctx context: %s")

	po.DropStaleTranslations()

	buff, err = po.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	po2 = NewPo()
	po2.Parse(buff)

	for k, v := range po.Headers {
		if v2, ok := po2.Headers[k]; ok {
			for i, value := range v {
				if value != v2[i] {
					t.Errorf("Only translations should have been dropped, not headers")
				}
			}
		}
	}

	tr = po2.Get("My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}
	tr = po2.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}

	tr = po2.Get("Some random")
	if tr == "Some random translation" || tr != "Some random" {
		t.Errorf("Expected 'Some random' translation to be dropped; was present")
	}

	// With 'the' removed?
	v = "Test"
	tr = po.GetC("One with var: %s", "Ctx", v)
	if tr != "This one is singular in a Ctx context: Test" {
		t.Errorf("Expected 'This one is singular in a Ctx context: Test' but got '%s'", tr)
	}

	tr = po.GetNC("One with var: %s", "Several with vars: %s", 17, "Ctx", v)
	if tr != "This one is plural in a Ctx context: Test" {
		t.Errorf("Expected 'This one is plural in a Ctx context: Test' but got '%s'", tr)
	}
}
