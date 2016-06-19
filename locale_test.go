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
msgstr "This one sets the var: %s"

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

	// Add domain
	l.AddDomain("my_domain")

	// Test translations
	tr := l.GetD("my_domain", "My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}

	v := "Variable"
	tr = l.GetD("my_domain", "One with var: %s", v)
	if tr != "This one sets the var: Variable" {
		t.Errorf("Expected 'This one sets the var: Variable' but got '%s'", tr)
	}
}
