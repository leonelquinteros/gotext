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

func TestMo_Get(t *testing.T) {
	// Create mo object
	mo := NewMo()

	// Try to parse a directory
	mo.ParseFile(path.Clean(os.TempDir()))

	// Parse file
	mo.ParseFile("fixtures/en_US/default.mo")

	// Test translations
	tr := mo.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = mo.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestMo(t *testing.T) {
	// Create mo object
	mo := NewMo()

	// Try to parse a directory
	mo.ParseFile(path.Clean(os.TempDir()))

	// Parse file
	mo.ParseFile("fixtures/en_US/default.mo")

	// Test translations
	tr := mo.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}

	v := "Variable"
	tr = mo.Get("One with var: %s", v)
	if tr != "This one is the singular: Variable" {
		t.Errorf("Expected 'This one is the singular: Variable' but got '%s'", tr)
	}

	// Test multi-line id
	tr = mo.Get("multilineid")
	if tr != "id with multiline content" {
		t.Errorf("Expected 'id with multiline content' but got '%s'", tr)
	}

	// Test multi-line plural id
	tr = mo.Get("multilinepluralid")
	if tr != "plural id with multiline content" {
		t.Errorf("Expected 'plural id with multiline content' but got '%s'", tr)
	}

	// Test multi-line
	tr = mo.Get("Multi-line")
	if tr != "Multi line" {
		t.Errorf("Expected 'Multi line' but got '%s'", tr)
	}

	// Test plural
	tr = mo.GetN("One with var: %s", "Several with vars: %s", 2, v)
	if tr != "This one is the plural: Variable" {
		t.Errorf("Expected 'This one is the plural: Variable' but got '%s'", tr)
	}

	// Test not existent translations
	tr = mo.Get("This is a test")
	if tr != "This is a test" {
		t.Errorf("Expected 'This is a test' but got '%s'", tr)
	}

	tr = mo.GetN("This is a test", "This are tests", 100)
	if tr != "This are tests" {
		t.Errorf("Expected 'This are tests' but got '%s'", tr)
	}

	// Test context translations
	v = "Test"
	tr = mo.GetC("One with var: %s", "Ctx", v)
	if tr != "This one is the singular in a Ctx context: Test" {
		t.Errorf("Expected 'This one is the singular in a Ctx context: Test' but got '%s'", tr)
	}

	// Test plural
	tr = mo.GetNC("One with var: %s", "Several with vars: %s", 17, "Ctx", v)
	if tr != "This one is the plural in a Ctx context: Test" {
		t.Errorf("Expected 'This one is the plural in a Ctx context: Test' but got '%s'", tr)
	}

	// Test default plural vs singular return responses
	tr = mo.GetN("Original", "Original plural", 4)
	if tr != "Original plural" {
		t.Errorf("Expected 'Original plural' but got '%s'", tr)
	}
	tr = mo.GetN("Original", "Original plural", 1)
	if tr != "Original" {
		t.Errorf("Expected 'Original' but got '%s'", tr)
	}

	// Test empty Translation strings
	tr = mo.Get("Empty Translation")
	if tr != "Empty Translation" {
		t.Errorf("Expected 'Empty Translation' but got '%s'", tr)
	}

	tr = mo.Get("Empty plural form singular")
	if tr != "Singular translated" {
		t.Errorf("Expected 'Singular translated' but got '%s'", tr)
	}

	tr = mo.GetN("Empty plural form singular", "Empty plural form", 1)
	if tr != "Singular translated" {
		t.Errorf("Expected 'Singular translated' but got '%s'", tr)
	}

	tr = mo.GetN("Empty plural form singular", "Empty plural form", 2)
	if tr != "Empty plural form" {
		t.Errorf("Expected 'Empty plural form' but got '%s'", tr)
	}

	// Test last Translation
	tr = mo.Get("More")
	if tr != "More translation" {
		t.Errorf("Expected 'More translation' but got '%s'", tr)
	}
}

func TestMoRace(t *testing.T) {
	// Create mo object
	mo := NewMo()

	// Create sync channels
	pc := make(chan bool)
	rc := make(chan bool)

	// Parse po content in a goroutine
	go func(mo *Mo, done chan bool) {
		// Parse file
		mo.ParseFile("fixtures/en_US/default.mo")
		done <- true
	}(mo, pc)

	// Read some Translation on a goroutine
	go func(mo *Mo, done chan bool) {
		mo.Get("My text")
		done <- true
	}(mo, rc)

	// Read something at top level
	mo.Get("My text")

	// Wait for goroutines to finish
	<-pc
	<-rc
}

func TestNewMoTranslatorRace(t *testing.T) {

	// Create Po object
	mo := NewMo()

	// Create sync channels
	pc := make(chan bool)
	rc := make(chan bool)

	// Parse po content in a goroutine
	go func(mo Translator, done chan bool) {
		// Parse file
		mo.ParseFile("fixtures/en_US/default.mo")
		done <- true
	}(mo, pc)

	// Read some Translation on a goroutine
	go func(mo Translator, done chan bool) {
		mo.Get("My text")
		done <- true
	}(mo, rc)

	// Read something at top level
	mo.Get("My text")

	// Wait for goroutines to finish
	<-pc
	<-rc
}

func TestMoBinaryEncoding(t *testing.T) {
	// Create mo objects
	mo := NewMo()
	mo2 := NewMo()

	// Parse file
	mo.ParseFile("fixtures/en_US/default.mo")

	buff, err := mo.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = mo2.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	// Test translations
	tr := mo2.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = mo2.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}
