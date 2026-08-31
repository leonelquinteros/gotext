/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
)

const (
	translatedText = "Translated text"
)

func TestPo_Get(t *testing.T) {
	var pos = []*Po{
		NewPo(), // test os
		NewPoFS(os.DirFS(".")),
	}

	for _, po := range pos {
		// Try to parse a directory
		po.ParseFile(path.Clean(os.TempDir()))

		// Parse file
		po.ParseFile("fixtures/en_US/default.po")

		// Test translations
		tr := po.Get("My text")
		if tr != translatedText {
			t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
		}

		v := "My text"
		tr = po.Get(v)
		if tr != translatedText {
			t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
		}

		// Test translations
		tr = po.Get("language")
		if tr != "en_US" {
			t.Errorf("Expected 'en_US' but got '%s'", tr)
		}
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
	defer func() {
		_ = f.Close()
	}()

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

func TestPoWrapperBehavior(t *testing.T) {
	po := NewPo()
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

func TestPoParseLinesBoundaries(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		po := NewPo()
		po.Parse(nil)

		translation, ok := po.GetDomain().translations[""]
		if !ok {
			t.Fatal("expected empty translation entry")
		}
		if translation == nil {
			t.Fatal("expected non-nil empty translation")
		}
		if len(translation.Trs) != 0 {
			t.Fatalf("expected no forms for empty input, got %v", translation.Trs)
		}
		if po.Headers == nil {
			t.Fatal("expected initialized headers for empty input")
		}
	})

	t.Run("unterminated final msgstr", func(t *testing.T) {
		po := NewPo()
		po.Parse([]byte("msgid \"id\"\nmsgstr \"translated\""))

		translation, ok := po.GetDomain().translations["id"]
		if !ok {
			t.Fatal("expected translation for unterminated final line")
		}
		if got := translation.Trs[0]; got != "translated" {
			t.Fatalf("translation = %q, want %q", got, "translated")
		}
	})
}

func TestPoParseInvalidMessageContinuations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid index",
			input: "msgid \"x\"\nmsgstr[bad]\n\"tail\"\n",
		},
		{
			name:  "invalid payload",
			input: "msgid \"x\"\nmsgstr[2] \"unterminated\n\"tail\"\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			po := NewPo()
			po.Parse([]byte(test.input))

			translation, ok := po.GetDomain().translations["x"]
			if !ok {
				t.Fatal("expected untranslated entry for msgid")
			}
			if len(translation.Trs) != 0 {
				t.Fatalf("invalid msgstr created forms: %v", translation.Trs)
			}
			if _, ok := po.GetDomain().translations["xtail"]; ok {
				t.Fatal("invalid msgstr continuation modified the msgid")
			}
		})
	}
}

func TestPoParseSparsePluralContinuation(t *testing.T) {
	po := NewPo()
	po.Parse([]byte("msgid \"x\"\nmsgstr[2] \"two\"\n\" tail\"\n"))

	translation, ok := po.GetDomain().translations["x"]
	if !ok {
		t.Fatal("expected plural translation")
	}
	if got := translation.Trs[2]; got != "two tail" {
		t.Fatalf("plural form 2 = %q, want %q", got, "two tail")
	}
	if _, ok := translation.Trs[0]; ok {
		t.Fatal("unexpected synthetic plural form 0")
	}
	if len(translation.Trs) != 1 {
		t.Fatalf("plural forms = %v, want only index 2", translation.Trs)
	}
}

func TestPoParseRejectsMalformedPluralIndexes(t *testing.T) {
	tests := []string{
		"",
		"-1",
		"+1",
		" 1",
		"1x",
		"١",
		"18446744073709551616",
	}

	for _, index := range tests {
		t.Run(fmt.Sprintf("index_%q", index), func(t *testing.T) {
			po := NewPo()
			po.Parse([]byte("msgid \"id\"\nmsgstr[" + index + "] \"form\"\n\"tail\"\n"))

			translation, ok := po.GetDomain().translations["id"]
			if !ok {
				t.Fatal("expected untranslated entry for msgid")
			}
			if len(translation.Trs) != 0 {
				t.Fatalf("malformed plural index created forms: %v", translation.Trs)
			}
			for form := range translation.Trs {
				if form < 0 {
					t.Fatalf("malformed plural index created negative form %d", form)
				}
			}
		})
	}
}

func TestPoParseRejectsInvalidDirectivesWithoutMutation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "msgid",
			input: "msgid \"unterminated\n\"tail\"\nmsgstr \"ignored\"\n",
		},
		{
			name:  "msgctxt",
			input: "msgctxt \"unterminated\n\"tail\"\n",
		},
		{
			name:  "msgid_plural",
			input: "msgid \"id\"\nmsgid_plural \"unterminated\n\"tail\"\nmsgstr[0] \"ignored\"\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			po := NewPo()
			po.Set("sentinel", "preserved")
			po.Parse([]byte(test.input))

			if got := po.Get("sentinel"); got != "preserved" {
				t.Fatalf("sentinel translation = %q, want %q", got, "preserved")
			}
			if _, ok := po.GetDomain().translations["tail"]; ok {
				t.Fatal("invalid directive continuation created an entry")
			}
			if test.name == "msgid_plural" {
				translation, ok := po.GetDomain().translations["id"]
				if !ok {
					t.Fatal("expected valid msgid to remain untranslated")
				}
				if len(translation.Trs) != 0 {
					t.Fatalf("invalid plural directive created forms: %v", translation.Trs)
				}
				if len(po.GetDomain().pluralTranslations) != 0 {
					t.Fatalf("plural translations = %v, want none", po.GetDomain().pluralTranslations)
				}
			}
			wantTranslations := 1
			if test.name == "msgid_plural" {
				wantTranslations = 2
			}
			if len(po.GetDomain().translations) != wantTranslations {
				t.Fatalf("translations = %v, want %d entries", po.GetDomain().translations, wantTranslations)
			}
		})
	}
}

func TestPoParseRequiresDirectiveKeywordBoundaries(t *testing.T) {
	for _, keyword := range []string{"msgctxt", "msgid", "msgid_plural", "msgstr"} {
		t.Run(keyword, func(t *testing.T) {
			po := NewPo()
			po.Set("sentinel", "preserved")
			po.Parse([]byte(keyword + "suffix \"value\"\n\"tail\"\n"))

			if got := po.Get("sentinel"); got != "preserved" {
				t.Fatalf("sentinel translation = %q, want %q", got, "preserved")
			}
			if len(po.GetDomain().translations) != 1 {
				t.Fatalf("translations = %v, want only the sentinel", po.GetDomain().translations)
			}
		})
	}
}

func TestPoParseEscapedMultilineContextIDPluralAndForms(t *testing.T) {
	po := NewPo()
	po.Parse([]byte(`msgctxt "ctx\"\\\n\x00"
"tail"
msgid "id\"\\\n\x00"
"tail"
msgid_plural "plural\"\\\n\x00"
"tail"
msgstr[0] "one\"\\\n\x00"
"tail"
msgstr[1] "two"
`))

	context := "ctx\"\\\n\x00tail"
	id := "id\"\\\n\x00tail"
	plural := "plural\"\\\n\x00tail"
	if _, ok := po.GetDomain().pluralTranslations[plural]; !ok {
		t.Fatal("expected completed plural ID in plural translation map")
	}
	translation, ok := po.GetDomain().contextTranslations[context][id]
	if !ok {
		t.Fatal("expected escaped context translation")
	}
	if got := translation.PluralID; got != plural {
		t.Fatalf("plural ID = %q, want %q", got, plural)
	}
	if got := translation.Trs[0]; got != "one\"\\\n\x00tail" {
		t.Fatalf("translation form 0 = %q, want %q", got, "one\"\\\n\x00tail")
	}
	if got := translation.Trs[1]; got != "two" {
		t.Fatalf("translation form 1 = %q, want %q", got, "two")
	}
}

func FuzzPoParseMalformedInput(f *testing.F) {
	f.Add([]byte("msgid \"id\"\nmsgstr \"translated\"\n"))
	f.Add([]byte("msgid \"id\"\nmsgstr[-1] \"bad\"\n\"tail\"\n"))
	f.Add([]byte("msgid \"unterminated\n\"tail\"\n"))
	f.Add([]byte("msgid_plural \"plural\"\nmsgstr[+1] \"bad\"\n"))
	f.Add([]byte("msgstring \"not a directive\"\n\"tail\"\n"))

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}

		po := NewPo()
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("Parse panicked: %v", recovered)
			}
		}()

		po.Parse(input)
		for id, translation := range po.GetDomain().translations {
			if translation == nil {
				t.Fatalf("translation %q is nil", id)
			}
			for form := range translation.Trs {
				if form < 0 {
					t.Fatalf("translation %q has negative plural form %d", id, form)
				}
			}
		}
		for context, translations := range po.GetDomain().contextTranslations {
			for id, translation := range translations {
				if translation == nil {
					t.Fatalf("context translation %q/%q is nil", context, id)
				}
				for form := range translation.Trs {
					if form < 0 {
						t.Fatalf("context translation %q/%q has negative plural form %d", context, id, form)
					}
				}
			}
		}
	})
}

func FuzzPoRejectedRecordDoesNotMutate(f *testing.F) {
	f.Add([]byte("rejected record"))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}

		po := NewPo()
		po.Set("ordinary id", "ordinary translation")
		po.SetC("contextual id", "context", "contextual translation")

		wantTranslations := po.GetDomain().GetTranslations()
		wantContextTranslations := po.GetDomain().GetCtxTranslations()

		quoted := strconv.Quote(string(input))
		record := []byte("msgid " + quoted[:len(quoted)-1])

		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("Parse panicked: %v", recovered)
			}
		}()

		po.Parse(record)

		if got := po.GetDomain().GetTranslations(); !reflect.DeepEqual(got, wantTranslations) {
			t.Fatalf("ordinary translations changed: got %#v, want %#v", got, wantTranslations)
		}
		if got := po.GetDomain().GetCtxTranslations(); !reflect.DeepEqual(got, wantContextTranslations) {
			t.Fatalf("context translations changed: got %#v, want %#v", got, wantContextTranslations)
		}
	})
}
