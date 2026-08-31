/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"encoding/binary"
	"os"
	"path"
	"testing"
)

func TestMo_Get(t *testing.T) {
	// Create mo object
	mos := []*Mo{
		NewMo(),
		NewMoFS(os.DirFS(".")),
	}

	for _, mo := range mos {
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

func TestMo_MissingWrappers(t *testing.T) {
	mo := NewMo()
	res := mo.Append(nil, "test")
	if string(res) != "test" {
		t.Error("Mo.Append failed")
	}

	res = mo.AppendN(nil, "one", "many", 1)
	if string(res) != "one" {
		t.Error("Mo.AppendN failed")
	}

	res = mo.AppendC(nil, "id", "ctx")
	if string(res) != "id" {
		t.Error("Mo.AppendC failed")
	}

	res = mo.AppendNC(nil, "id", "plural", 1, "ctx")
	if string(res) != "id" {
		t.Error("Mo.AppendNC failed")
	}

	if mo.IsTranslated("id") {
		t.Error("Mo.IsTranslated failed")
	}
	if mo.IsTranslatedN("id", 1) {
		t.Error("Mo.IsTranslatedN failed")
	}
	if mo.IsTranslatedC("id", "ctx") {
		t.Error("Mo.IsTranslatedC failed")
	}
	if mo.IsTranslatedNC("id", 1, "ctx") {
		t.Error("Mo.IsTranslatedNC failed")
	}
}

func TestMoParseRejectsHugeCount(t *testing.T) {
	buf := make([]byte, 28)
	binary.LittleEndian.PutUint32(buf[0:4], MoMagicLittleEndian)
	binary.LittleEndian.PutUint32(buf[8:12], ^uint32(0))

	mo := NewMo()
	mo.GetDomain().Set("seed", "preserved")
	mo.Parse(buf)

	if got := mo.Get("seed"); got != "preserved" {
		t.Fatalf("seed translation = %q, want %q", got, "preserved")
	}
	if len(mo.domain.translations) != 1 {
		t.Fatalf("translations = %v, want only the seed", mo.domain.translations)
	}
}

func TestMoParseRejectsInvalidPayloadWithoutMutation(t *testing.T) {
	t.Run("payload extends past file", func(t *testing.T) {
		buf := make([]byte, 46)
		binary.LittleEndian.PutUint32(buf[0:4], MoMagicLittleEndian)
		binary.LittleEndian.PutUint32(buf[8:12], 1)
		binary.LittleEndian.PutUint32(buf[12:16], 28)
		binary.LittleEndian.PutUint32(buf[16:20], 36)
		binary.LittleEndian.PutUint32(buf[28:32], 4)
		binary.LittleEndian.PutUint32(buf[32:36], 44)
		binary.LittleEndian.PutUint32(buf[36:40], 1)
		binary.LittleEndian.PutUint32(buf[40:44], 45)
		copy(buf[44:], "xy")

		mo := NewMo()
		mo.GetDomain().Set("seed", "preserved")
		mo.Parse(buf)

		if got := mo.Get("seed"); got != "preserved" {
			t.Fatalf("seed translation = %q, want %q", got, "preserved")
		}
		if len(mo.domain.translations) != 1 {
			t.Fatalf("translations = %v, want only the seed", mo.domain.translations)
		}
	})

	t.Run("later payload invalidates all entries", func(t *testing.T) {
		buf := make([]byte, 70)
		binary.LittleEndian.PutUint32(buf[0:4], MoMagicLittleEndian)
		binary.LittleEndian.PutUint32(buf[8:12], 2)
		binary.LittleEndian.PutUint32(buf[12:16], 28)
		binary.LittleEndian.PutUint32(buf[16:20], 44)

		binary.LittleEndian.PutUint32(buf[28:32], 3)
		binary.LittleEndian.PutUint32(buf[32:36], 60)
		binary.LittleEndian.PutUint32(buf[36:40], 3)
		binary.LittleEndian.PutUint32(buf[40:44], 66)

		binary.LittleEndian.PutUint32(buf[44:48], 3)
		binary.LittleEndian.PutUint32(buf[48:52], 63)
		binary.LittleEndian.PutUint32(buf[52:56], 3)
		binary.LittleEndian.PutUint32(buf[56:60], 69)
		copy(buf[60:], "oneONEtwoT")

		mo := NewMo()
		mo.GetDomain().Set("seed", "preserved")
		mo.Parse(buf)

		if got := mo.Get("seed"); got != "preserved" {
			t.Fatalf("seed translation = %q, want %q", got, "preserved")
		}
		if len(mo.domain.translations) != 1 {
			t.Fatalf("translations = %v, want only the seed", mo.domain.translations)
		}
	})
}

func TestMoAddTranslationPreservesPluralBytes(t *testing.T) {
	mo := NewMo()
	mo.addTranslation(
		[]byte("id\x00plural\x00extra"),
		[]byte("one\x00two\x00three"),
	)

	translation, ok := mo.domain.translations["id"]
	if !ok {
		t.Fatal("expected translation for repeated-NUL msgid")
	}
	if got := translation.PluralID; got != "plural\x00extra" {
		t.Fatalf("plural ID = %q, want %q", got, "plural\x00extra")
	}
	for i, want := range map[int]string{0: "one", 1: "two", 2: "three"} {
		if got := translation.Trs[i]; got != want {
			t.Fatalf("translation form %d = %q, want %q", i, got, want)
		}
	}
	if len(translation.Trs) != 3 {
		t.Fatalf("translation forms = %v, want 3 forms", translation.Trs)
	}

	mo = NewMo()
	mo.addTranslation([]byte("singular"), []byte("translated"))
	translation = mo.domain.translations["singular"]
	if got := translation.PluralID; got != "" {
		t.Fatalf("singular plural ID = %q, want empty", got)
	}
}
