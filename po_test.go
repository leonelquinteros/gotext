package gotext

import (
	"os"
	"path"
	"testing"
)

func TestPo(t *testing.T) {
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

	// Parse po file
	po := new(Po)
	po.ParseFile(filename)

	// Test translations
	tr := po.Get("My text")
	if tr != "Translated text" {
		t.Errorf("Expected 'Translated text' but got '%s'", tr)
	}

	v := "Variable"
	tr = po.Get("One with var: %s", v)
	if tr != "This one is the singular: Variable" {
		t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
	}

	// Test plural
	tr = po.GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "And this is the second plural form: Variable" {
		t.Errorf("Expected 'And this is the second plural form: Variable' but got '%s'", tr)
	}
}
