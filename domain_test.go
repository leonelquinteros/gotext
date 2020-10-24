package gotext

import (
	"testing"

	"golang.org/x/text/feature/plural"
)

//since both Po and Mo just pass-through to Domain for MarshalBinary and UnmarshalBinary, test it here
func TestBinaryEncoding(t *testing.T) {
	// Create po objects
	po := NewPo()
	po2 := NewPo()

	// Parse file
	po.ParseFile("fixtures/en_US/default.po")

	buff, err := po.GetDomain().MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = po2.GetDomain().UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	// Test translations
	tr := po2.Get("My text")
	if tr != translatedText {
		t.Errorf("Expected '%s' but got '%s'", translatedText, tr)
	}
	// Test translations
	tr = po2.Get("language")
	if tr != "en_US" {
		t.Errorf("Expected 'en_US' but got '%s'", tr)
	}
}

func TestGetAll(t *testing.T) {
	po := NewPo()
	po.ParseFile("fixtures/en_US/default.po")

	all, err := po.GetDomain().GetAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(all) < 5 {
		t.Errorf("did not get enough translations: only got %d", len(all))
	}

	const msgID = "My text"
	const msgStr = "Translated text"

	trans, ok := all[msgID]
	if !ok {
		t.Errorf("could not find expected item: msgid %s", msgID)
	}
	if trans.ID != msgID {
		t.Error("translation ID not as expected")
	}
	if trans.String != msgStr {
		t.Error("translation String not as expected")
	}
	if trans.Get() != msgStr {
		t.Error("translation.Get() not as expected")
	}

	const pluralMsgID = "One with var: %s"
	const pluralID = "Several with vars: %s"
	const pluralStr0 = "This one is the singular: %s"
	const pluralStr1 = "This one is the plural: %s"

	trans, ok = all[pluralMsgID]
	if !ok {
		t.Errorf("could not find expected item: msgid %s", pluralMsgID)
	}
	plTrans, ok := all[pluralID]
	if !ok {
		t.Errorf("could not find expected item: pluralid %s", pluralID)
	}
	if trans != plTrans {
		t.Error("pluralMsgID is not equal to pluralID")
	}
	if len(trans.Plurals) != 2 {
		t.Errorf("expected 2 plural forms but got %d", len(trans.Plurals))
	}
	if trans.GetPlural(plural.One) != pluralStr0 {
		t.Errorf("plural form One expected \"%s\" but got \"%s\"", pluralStr0, trans.GetPlural(plural.One))
	}
	if trans.GetPlural(plural.Other) != pluralStr1 {
		t.Errorf("plural form Other expected \"%s\" but got \"%s\"", pluralStr1, trans.GetPlural(plural.Other))
	}
}
