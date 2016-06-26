package gotext

import (
	"os"
	"path"
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
    
    // Set default configuration
    Configure("/tmp", "en_US", "default")
    
	// Create Locales directory on default location
	dirname := path.Clean(library + string(os.PathSeparator) + "en_US")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to default domain file
	filename := path.Clean(dirname + string(os.PathSeparator) + domain + ".po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Test translations
	tr := Get("My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}

	v := "Variable"
	tr = Get("One with var: %s", v)
	if tr != "This one is the singular: Variable" {
		t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
	}

	// Test plural
	tr = GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "And this is the second plural form: Variable" {
		t.Errorf("Expected 'And this is the second plural form: Variable' but got '%s'", tr)
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

    `

	// Create Locales directory on default location
	dirname := path.Clean(library + string(os.PathSeparator) + "en_US")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to default domain file
	filename := path.Clean(dirname + string(os.PathSeparator) + domain + ".po")

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Can't create test file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		t.Fatalf("Can't write to test file: %s", err.Error())
	}

	// Init sync channels
	c1 := make(chan bool)
	c2 := make(chan bool)

	// Test translations
	go func(done chan bool) {
		Get("My text")
		done <- true
	}(c1)

	go func(done chan bool) {
		Get("My text")
		done <- true
	}(c2)

	Get("My text")
}
