package gotext

import "testing"

func TestTranslation_NewTranslationWithRefs(t *testing.T) {
	refs := []string{"a.go:1", "b.go:2"}
	tr := NewTranslationWithRefs(refs)
	if len(tr.Refs) != 2 || tr.Refs[0] != "a.go:1" {
		t.Error("NewTranslationWithRefs failed")
	}
}

func TestTranslation_IsTranslated(t *testing.T) {
	tr := NewTranslation()
	if tr.IsTranslated() {
		t.Error("Expected false for empty translation")
	}
}
