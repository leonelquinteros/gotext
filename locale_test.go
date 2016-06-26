package gotext

import (
	"os"
	"path"
	"testing"
)

func TestLocale(t *testing.T) {
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
	dirname := path.Clean("/tmp" + string(os.PathSeparator) + "en")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to file
	filename := path.Clean(dirname + string(os.PathSeparator) + "my_domain.po")

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
	l.domains = nil

	// Add domain
	l.AddDomain("my_domain")

	// Test translations
	tr := l.GetD("my_domain", "My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}

	v := "Variable"
	tr = l.GetD("my_domain", "One with var: %s", v)
	if tr != "This one is the singular: Variable" {
		t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
	}

	// Test plural
	tr = l.GetND("my_domain", "One with var: %s", "Several with vars: %s", 2, v)
	if tr != "And this is the second plural form: Variable" {
		t.Errorf("Expected 'And this is the second plural form: Variable' but got '%s'", tr)
	}

	// Test non-existent "deafult" domain responses
	tr = l.Get("My text")
	if tr != "My text" {
		t.Errorf("Expected 'My text' but got '%s'", tr)
	}

	tr = l.GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "Several with vars: Variable" {
		t.Errorf("Expected 'Several with vars: Variable' but got '%s'", tr)
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
	dirname := path.Clean("/tmp" + string(os.PathSeparator) + "es")
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		t.Fatalf("Can't create test directory: %s", err.Error())
	}

	// Write PO content to file
	filename := path.Clean(dirname + string(os.PathSeparator) + "race.po")

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
