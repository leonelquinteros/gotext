/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"encoding/binary"
	"fmt"
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

type moFixtureEntry struct {
	msgid  []byte
	msgstr []byte
}

func makeMoFixture(order binary.ByteOrder, entries ...moFixtureEntry) []byte {
	const headerSize = 28
	directorySize := len(entries) * 8
	msgIDOffset := headerSize
	msgStrOffset := msgIDOffset + directorySize
	dataOffset := msgStrOffset + directorySize

	dataSize := 0
	for _, entry := range entries {
		dataSize += len(entry.msgid) + len(entry.msgstr)
	}
	buf := make([]byte, dataOffset+dataSize)

	order.PutUint32(buf[0:4], MoMagicLittleEndian)
	order.PutUint32(buf[8:12], uint32(len(entries)))
	order.PutUint32(buf[12:16], uint32(msgIDOffset))
	order.PutUint32(buf[16:20], uint32(msgStrOffset))

	cursor := dataOffset
	for i, entry := range entries {
		idEntry := msgIDOffset + i*8
		strEntry := msgStrOffset + i*8
		order.PutUint32(buf[idEntry:idEntry+4], uint32(len(entry.msgid)))
		order.PutUint32(buf[idEntry+4:idEntry+8], uint32(cursor))
		copy(buf[cursor:], entry.msgid)
		cursor += len(entry.msgid)

		order.PutUint32(buf[strEntry:strEntry+4], uint32(len(entry.msgstr)))
		order.PutUint32(buf[strEntry+4:strEntry+8], uint32(cursor))
		copy(buf[cursor:], entry.msgstr)
		cursor += len(entry.msgstr)
	}

	return buf
}

func TestMoParseBigEndianSingularContextPlural(t *testing.T) {
	catalog := makeMoFixture(
		binary.BigEndian,
		moFixtureEntry{
			msgid:  []byte(""),
			msgstr: []byte("Language: xx\nPlural-Forms: nplurals=2; plural=(n != 1);\n"),
		},
		moFixtureEntry{
			msgid:  []byte("hello"),
			msgstr: []byte("bonjour"),
		},
		moFixtureEntry{
			msgid:  []byte("menu\x04title"),
			msgstr: []byte("Title"),
		},
		moFixtureEntry{
			msgid:  []byte("menu\x04item\x00items"),
			msgstr: []byte("one\x00many"),
		},
	)

	mo := NewMo()
	mo.Parse(catalog)

	if got := mo.Get("hello"); got != "bonjour" {
		t.Fatalf("singular lookup = %q, want %q", got, "bonjour")
	}
	if got := mo.GetC("title", "menu"); got != "Title" {
		t.Fatalf("context lookup = %q, want %q", got, "Title")
	}

	translation, ok := mo.GetDomain().contextTranslations["menu"]["item"]
	if !ok {
		t.Fatal("expected big-endian context translation")
	}
	if got := translation.PluralID; got != "items" {
		t.Fatalf("plural ID = %q, want %q", got, "items")
	}
	if got := translation.Trs[0]; got != "one" {
		t.Fatalf("translation form 0 = %q, want %q", got, "one")
	}
	if got := translation.Trs[1]; got != "many" {
		t.Fatalf("translation form 1 = %q, want %q", got, "many")
	}
	if got := mo.GetNC("item", "items", 2, "menu"); got != "many" {
		t.Fatalf("plural context lookup = %q, want %q", got, "many")
	}
}

func TestMoParseRejectsInvalidMagicWithoutMutation(t *testing.T) {
	catalog := makeMoFixture(
		binary.BigEndian,
		moFixtureEntry{msgid: []byte("id"), msgstr: []byte("translated")},
	)
	copy(catalog[:4], []byte{0x95, 0x04, 0x12, 0xdf})

	mo := NewMo()
	mo.GetDomain().Set("sentinel", "preserved")
	mo.Parse(catalog)

	if got := mo.Get("sentinel"); got != "preserved" {
		t.Fatalf("sentinel translation = %q, want %q", got, "preserved")
	}
	if len(mo.GetDomain().translations) != 1 {
		t.Fatalf("translations = %v, want only the sentinel", mo.GetDomain().translations)
	}
}

func TestMoAddTranslationPreservesEotSuffix(t *testing.T) {
	mo := NewMo()
	mo.addTranslation(
		[]byte("menu\x04item\x04suffix\x00items"),
		[]byte("one\x00many"),
	)

	translation, ok := mo.GetDomain().contextTranslations["menu"]["item\x04suffix"]
	if !ok {
		t.Fatal("expected context translation with EOT suffix")
	}
	if got := translation.ID; got != "item\x04suffix" {
		t.Fatalf("translation ID = %q, want %q", got, "item\\x04suffix")
	}
	if got := translation.PluralID; got != "items" {
		t.Fatalf("plural ID = %q, want %q", got, "items")
	}
}

func TestMoParseTruncatedHeaderPreservesSeed(t *testing.T) {
	catalog := makeMoFixture(
		binary.LittleEndian,
		moFixtureEntry{msgid: []byte("id"), msgstr: []byte("translated")},
	)

	for size := 0; size <= 27; size++ {
		t.Run(fmt.Sprintf("truncate_%d", size), func(t *testing.T) {
			mo := NewMo()
			mo.GetDomain().Set("sentinel", "preserved")
			mo.Parse(catalog[:size])

			if got := mo.Get("sentinel"); got != "preserved" {
				t.Fatalf("sentinel translation = %q, want %q", got, "preserved")
			}
			if len(mo.GetDomain().translations) != 1 {
				t.Fatalf("translations = %v, want only the sentinel", mo.GetDomain().translations)
			}
		})
	}
}

func TestMoParseRejectsDirectoryOneByteOverflow(t *testing.T) {
	buf := make([]byte, 35)
	binary.LittleEndian.PutUint32(buf[0:4], MoMagicLittleEndian)
	binary.LittleEndian.PutUint32(buf[8:12], 1)
	binary.LittleEndian.PutUint32(buf[12:16], 28)
	binary.LittleEndian.PutUint32(buf[16:20], 28)

	mo := NewMo()
	mo.GetDomain().Set("sentinel", "preserved")
	mo.Parse(buf)

	if got := mo.Get("sentinel"); got != "preserved" {
		t.Fatalf("sentinel translation = %q, want %q", got, "preserved")
	}
	if len(mo.GetDomain().translations) != 1 {
		t.Fatalf("translations = %v, want only the sentinel", mo.GetDomain().translations)
	}
}

func TestMoParseRejectsNearMaxUint32Offsets(t *testing.T) {
	base := makeMoFixture(
		binary.LittleEndian,
		moFixtureEntry{msgid: []byte("id"), msgstr: []byte("translated")},
	)
	maxUint32 := ^uint32(0)
	tests := []struct {
		name   string
		mutate func([]byte)
	}{
		{
			name: "msgid directory offset",
			mutate: func(buf []byte) {
				binary.LittleEndian.PutUint32(buf[12:16], maxUint32)
			},
		},
		{
			name: "msgstr directory offset",
			mutate: func(buf []byte) {
				binary.LittleEndian.PutUint32(buf[16:20], maxUint32)
			},
		},
		{
			name: "msgid payload offset",
			mutate: func(buf []byte) {
				binary.LittleEndian.PutUint32(buf[32:36], maxUint32)
			},
		},
		{
			name: "msgstr payload offset",
			mutate: func(buf []byte) {
				binary.LittleEndian.PutUint32(buf[40:44], maxUint32)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := append([]byte(nil), base...)
			test.mutate(catalog)

			mo := NewMo()
			mo.GetDomain().Set("sentinel", "preserved")
			mo.Parse(catalog)

			if got := mo.Get("sentinel"); got != "preserved" {
				t.Fatalf("sentinel translation = %q, want %q", got, "preserved")
			}
			if len(mo.GetDomain().translations) != 1 {
				t.Fatalf("translations = %v, want only the sentinel", mo.GetDomain().translations)
			}
		})
	}
}

func FuzzMoParseBoundedCatalog(f *testing.F) {
	littleEndianCatalog := makeMoFixture(
		binary.LittleEndian,
		moFixtureEntry{msgid: []byte("id"), msgstr: []byte("translated")},
	)
	bigEndianCatalog := makeMoFixture(
		binary.BigEndian,
		moFixtureEntry{msgid: []byte("id"), msgstr: []byte("translated")},
	)

	f.Add(littleEndianCatalog)
	f.Add(bigEndianCatalog)
	f.Add([]byte{})
	f.Add([]byte("not an MO catalog"))
	f.Add(littleEndianCatalog[:27])
	f.Add([]byte{0xde, 0x12, 0x04, 0x95, 0, 0, 0, 0})

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}

		mo := NewMo()
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("Parse panicked: %v", recovered)
			}
		}()

		mo.Parse(input)
		for id, translation := range mo.GetDomain().translations {
			if translation == nil {
				t.Fatalf("translation %q is nil", id)
			}
		}
		for context, translations := range mo.GetDomain().contextTranslations {
			for id, translation := range translations {
				if translation == nil {
					t.Fatalf("context translation %q/%q is nil", context, id)
				}
			}
		}
	})
}
