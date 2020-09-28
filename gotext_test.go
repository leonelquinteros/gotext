package gotext

import (
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"
)

func TestGettersSetters(t *testing.T) {
	SetDomain("test")
	dom := GetDomain()

	if dom != "test" {
		t.Errorf("Expected GetDomain to return 'test', but got '%s'", dom)
	}

	SetLibrary("/tmp/test")
	lib := GetLibrary()

	if lib != "/tmp/test" {
		t.Errorf("Expected GetLibrary to return '/tmp/test', but got '%s'", lib)
	}

	SetLanguage("es")
	lang := GetLanguage()

	if lang != "es" {
		t.Errorf("Expected GetLanguage to return 'es', but got '%s'", lang)
	}
}

func TestPackageFunctions(t *testing.T) {
	// Set PO content
	str := `
msgid   ""
msgstr  "Project-Id-Version: %s\n"
        "Report-Msgid-Bugs-To: %s\n"
        
# Initial comment
# More Headers below
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

msgid "Untranslated"
msgid_plural "Several untranslated"
msgstr[0] ""
msgstr[1] ""

	`

	anotherStr := `
msgid   ""
msgstr  "Project-Id-Version: %s\n"
        "Report-Msgid-Bugs-To: %s\n"

# Initial comment
# More Headers below
"Language: en\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Another text on a different domain"
msgstr "Another text on another domain"
  `

	// Create Locales directory on default location
	dirname := path.Join("/tmp", "en_US")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to default domain file
	filename := path.Join(dirname, "default.po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	anotherFilename := path.Join(dirname, "another.po")

	af, err := os.Create(anotherFilename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}

	_, err = af.WriteString(anotherStr)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Move file close to write the file, so we can use it in the next step
	f.Close()
	af.Close()

	// Set package configuration
	Configure("/tmp", "en_US", "default")

	// Test translations
	tr := Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}

	v := "Variable"
	tr = Get("One with var: %s", v)
	if tr != "This one is the singular: Variable" {
		t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
	}

	// Test plural
	tr = GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "This one is the plural: Variable" {
		t.Errorf("Expected 'This one is the plural: Variable' but got '%s'", tr)
	}

	// Test context translations
	tr = GetC("Some random in a context", "Ctx")
	if tr != "Some random Translation in a context" {
		t.Errorf("Expected 'Some random Translation in a context' but got '%s'", tr)
	}

	v = "Variable"
	tr = GetC("One with var: %s", "Ctx", v)
	if tr != "This one is the singular in a Ctx context: Variable" {
		t.Errorf("Expected 'This one is the singular in a Ctx context: Variable' but got '%s'", tr)
	}

	tr = GetNC("One with var: %s", "Several with vars: %s", 19, "Ctx", v)
	if tr != "This one is the plural in a Ctx context: Variable" {
		t.Errorf("Expected 'This one is the plural in a Ctx context: Variable' but got '%s'", tr)
	}

	tr = GetD("another", "Another text on a different domain")
	if tr != "Another text on another domain" {
		t.Errorf("Expected 'Another text on another domain' but got '%s'", tr)
	}
}

func TestUntranslated(t *testing.T) {
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

msgid "Untranslated"
msgid_plural "Several untranslated"
msgstr[0] ""
msgstr[1] ""

	`

	// Create Locales directory on default location
	dirname := path.Join("/tmp", "en_US")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to default domain file
	filename := path.Join(dirname, "default.po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Set package configuration
	Configure("/tmp", "en_US", "default")

	// Test untranslated
	tr := Get("Untranslated")
	if tr != "Untranslated" {
		t.Errorf("Expected 'Untranslated' but got '%s'", tr)
	}
	tr = GetN("Untranslated", "Several untranslated", 1)
	if tr != "Untranslated" {
		t.Errorf("Expected 'Untranslated' but got '%s'", tr)
	}

	tr = GetN("Untranslated", "Several untranslated", 2)
	if tr != "Several untranslated" {
		t.Errorf("Expected 'Several untranslated' but got '%s'", tr)
	}

	tr = GetD("default", "Untranslated")
	if tr != "Untranslated" {
		t.Errorf("Expected 'Untranslated' but got '%s'", tr)
	}
	tr = GetND("default", "Untranslated", "Several untranslated", 1)
	if tr != "Untranslated" {
		t.Errorf("Expected 'Untranslated' but got '%s'", tr)
	}

	tr = GetND("default", "Untranslated", "Several untranslated", 2)
	if tr != "Several untranslated" {
		t.Errorf("Expected 'Several untranslated' but got '%s'", tr)
	}
}

func TestMoAndPoTranslator(t *testing.T) {
	fixPath, _ := filepath.Abs("./fixtures/")

	Configure(fixPath, "en_GB", "default")

	// Check default domain Translation
	SetDomain("default")
	tr := Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s'. Got '%s'", translatedText, tr)
	}
	tr = Get("language")
	if tr != "en_GB" {
		t.Errorf("Expected 'en_GB'. Got '%s'", tr)
	}

	// Change Language (locale)
	SetLanguage("en_AU")

	// Check default domain Translation
	SetDomain("default")
	tr = Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s'. Got '%s'", translatedText, tr)
	}
	tr = Get("language")
	if tr != "en_AU" {
		t.Errorf("Expected 'en_AU'. Got '%s'", tr)
	}
}

func TestDomains(t *testing.T) {
	// Set PO content
	strDefault := `
msgid ""
msgstr "Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Default text"
msgid_plural "Default texts"
msgstr[0] "Default Translation"
msgstr[1] "Default translations"

msgctxt "Ctx"
msgid "Default context"
msgid_plural "Default contexts"
msgstr[0] "Default ctx Translation"
msgstr[1] "Default ctx translations"
	`

	strCustom := `
msgid ""
msgstr "Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Custom text"
msgid_plural "Custom texts"
msgstr[0] "Custom Translation"
msgstr[1] "Custom translations"

msgctxt "Ctx"
msgid "Custom context"
msgid_plural "Custom contexts"
msgstr[0] "Custom ctx Translation"
msgstr[1] "Custom ctx translations"
	`

	// Create Locales directory and files on temp location
	dirname := path.Join("/tmp", "en_US")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	fDefault, err := os.Create(path.Join(dirname, "default.po"))
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer fDefault.Close()

	fCustom, err := os.Create(path.Join(dirname, "custom.po"))
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer fCustom.Close()

	_, err = fDefault.WriteString(strDefault)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}
	_, err = fCustom.WriteString(strCustom)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	Configure("/tmp", "en_US", "default")

	// Check default domain Translation
	SetDomain("default")
	tr := Get("Default text")
	if tr != "Default Translation" {
		t.Errorf("Expected 'Default Translation'. Got '%s'", tr)
	}
	tr = GetN("Default text", "Default texts", 23)
	if tr != "Default translations" {
		t.Errorf("Expected 'Default translations'. Got '%s'", tr)
	}
	tr = GetC("Default context", "Ctx")
	if tr != "Default ctx Translation" {
		t.Errorf("Expected 'Default ctx Translation'. Got '%s'", tr)
	}
	tr = GetNC("Default context", "Default contexts", 23, "Ctx")
	if tr != "Default ctx translations" {
		t.Errorf("Expected 'Default ctx translations'. Got '%s'", tr)
	}

	SetDomain("custom")
	tr = Get("Custom text")
	if tr != "Custom Translation" {
		t.Errorf("Expected 'Custom Translation'. Got '%s'", tr)
	}
	tr = GetN("Custom text", "Custom texts", 23)
	if tr != "Custom translations" {
		t.Errorf("Expected 'Custom translations'. Got '%s'", tr)
	}
	tr = GetC("Custom context", "Ctx")
	if tr != "Custom ctx Translation" {
		t.Errorf("Expected 'Custom ctx Translation'. Got '%s'", tr)
	}
	tr = GetNC("Custom context", "Custom contexts", 23, "Ctx")
	if tr != "Custom ctx translations" {
		t.Errorf("Expected 'Custom ctx translations'. Got '%s'", tr)
	}
}

func TestPackageRace(t *testing.T) {
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

msgctxt "Ctx"
msgid "Some random in a context"
msgstr "Some random Translation in a context"

	`

	// Create Locales directory on default location
	dirname := path.Join("/tmp", "en_US")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to default domain file
	filename := path.Join("/tmp", GetDomain()+".po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		// Test translations
		go func() {
			defer wg.Done()

			GetLibrary()
			SetLibrary(path.Join("/tmp", "gotextlib"))
			GetDomain()
			SetDomain("default")
			GetLanguage()
			SetLanguage("en_US")
			Configure("/tmp", "en_US", "default")

			Get("My text")
			GetN("One with var: %s", "Several with vars: %s", 0, "test")
			GetC("Some random in a context", "Ctx")
		}()
	}

	wg.Wait()
}

func TestPackageArabicTranslation(t *testing.T) {
	Configure("fixtures/", "ar", "categories")

	// Plurals formula missing + Plural translation string missing
	tr := GetD("categories", "Alcohol & Tobacco")
	if tr != "الكحول والتبغ" {
		t.Errorf("Expected to get 'الكحول والتبغ', but got '%s'", tr)
	}

	// Plural translation string present without translations, should get the msgid_plural
	tr = GetND("categories", "%d selected", "%d selected", 10)
	if tr != "%d selected" {
		t.Errorf("Expected to get '%%d selected', but got '%s'", tr)
	}

	//Plurals formula present + Plural translation string present and complete
	tr = GetND("categories", "Load %d more document", "Load %d more documents", 0)
	if tr != "حمّل %d مستندات إضافيّة" {
		t.Errorf("Expected to get 'msgstr[0]', but got '%s'", tr)
	}

	tr = GetND("categories", "Load %d more document", "Load %d more documents", 1)
	if tr != "حمّل مستند واحد إضافي" {
		t.Errorf("Expected to get 'msgstr[1]', but got '%s'", tr)
	}

	tr = GetND("categories", "Load %d more document", "Load %d more documents", 2)
	if tr != "حمّل مستندين إضافيين" {
		t.Errorf("Expected to get 'msgstr[2]', but got '%s'", tr)
	}

	tr = GetND("categories", "Load %d more document", "Load %d more documents", 6)
	if tr != "حمّل %d مستندات إضافيّة" {
		t.Errorf("Expected to get 'msgstr[3]', but got '%s'", tr)
	}

	tr = GetND("categories", "Load %d more document", "Load %d more documents", 116)
	if tr != "حمّل %d مستندا إضافيّا" {
		t.Errorf("Expected to get 'msgstr[4]', but got '%s'", tr)
	}

	tr = GetND("categories", "Load %d more document", "Load %d more documents", 102)
	if tr != "حمّل %d مستند إضافي" {
		t.Errorf("Expected to get 'msgstr[5]', but got '%s'", tr)
	}
}

func TestPackageArabicMissingPluralForm(t *testing.T) {
	Configure("fixtures/", "ar", "no_plural_header")

	// Get translation
	tr := GetD("no_plural_header", "Alcohol & Tobacco")
	if tr != "الكحول والتبغ" {
		t.Errorf("Expected to get 'الكحول والتبغ', but got '%s'", tr)
	}
}
