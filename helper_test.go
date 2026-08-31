/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"reflect"
	"testing"
)

func TestSimplifiedLocale(t *testing.T) {
	tr := SimplifiedLocale("de_DE@euro")
	if tr != "de_DE" {
		t.Errorf("Expected 'de_DE' but got '%s'", tr)
	}

	tr = SimplifiedLocale("de_DE.UTF-8")
	if tr != "de_DE" {
		t.Errorf("Expected 'de_DE' but got '%s'", tr)
	}

	tr = SimplifiedLocale("de_DE:latin1")
	if tr != "de_DE" {
		t.Errorf("Expected 'de_DE' but got '%s'", tr)
	}
}

func TestReformattingSingleNamedPattern(t *testing.T) {
	pat := "%(name_me)x"

	f, n := reformatSprintf(pat)

	if f != "%x" {
		t.Errorf("pattern should be %%x but %v", f)
	}

	if !reflect.DeepEqual(n, []string{"name_me"}) {
		t.Errorf("named var should be {name_me} but %v", n)
	}
}

func TestReformattingMultipleNamedPattern(t *testing.T) {
	pat := "%(name_me)x and %(another_name)v"

	f, n := reformatSprintf(pat)

	if f != "%x and %v" {
		t.Errorf("pattern should be %%x and %%v but %v", f)
	}

	if !reflect.DeepEqual(n, []string{"name_me", "another_name"}) {
		t.Errorf("named var should be {name_me, another_name} but %v", n)
	}
}

func TestReformattingRepeatedNamedPattern(t *testing.T) {
	pat := "%(name_me)x and %(another_name)v and %(name_me)v"

	f, n := reformatSprintf(pat)

	if f != "%x and %v and %v" {
		t.Errorf("pattern should be %%x and %%v and %%v but %v", f)
	}

	if !reflect.DeepEqual(n, []string{"name_me", "another_name", "name_me"}) {
		t.Errorf("named var should be {name_me, another_name, name_me} but %v", n)
	}
}

func TestSprintf(t *testing.T) {
	pat := "%(brother)s loves %(sister)s. %(sister)s also loves %(brother)s."
	params := map[string]any{
		"sister":  "Susan",
		"brother": "Louis",
	}

	s := Sprintf(pat, params)

	if s != "Louis loves Susan. Susan also loves Louis." {
		t.Errorf("result should be Louis loves Susan. Susan also love Louis. but %v", s)
	}
}

func TestNPrintf(t *testing.T) {
	pat := "%(brother)s loves %(sister)s. %(sister)s also loves %(brother)s.\n"
	params := map[string]any{
		"sister":  "Susan",
		"brother": "Louis",
	}

	NPrintf(pat, params)

}

func TestSprintfFloatsWithPrecision(t *testing.T) {
	pat := "%(float)f / %(floatprecision).1f / %(long)g / %(longprecision).3g"
	params := map[string]any{
		"float":          5.034560,
		"floatprecision": 5.03456,
		"long":           5.03456,
		"longprecision":  5.03456,
	}

	s := Sprintf(pat, params)

	expectedresult := "5.034560 / 5.0 / 5.03456 / 5.03"
	if s != expectedresult {
		t.Errorf("result should be (%v) but is (%v)", expectedresult, s)
	}
}

func TestHelper_Appendf(t *testing.T) {
	b := []byte("test: ")
	res := Appendf(b, "Hello %s", "World")
	if string(res) != "test: Hello World" {
		t.Errorf("Expected 'test: Hello World', got '%s'", string(res))
	}
}

func TestReformatSprintfEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		wantFormat string
		wantNames  []string
	}{
		{
			name:       "repeated names",
			format:     "%(name)s %(name)q",
			wantFormat: "%s %q",
			wantNames:  []string{"name", "name"},
		},
		{
			name:       "width and precision",
			format:     "%(amount)08.2f",
			wantFormat: "%08.2f",
			wantNames:  []string{"amount"},
		},
		{
			name:       "adjacent placeholders",
			format:     "%(left)s%(right)d",
			wantFormat: "%s%d",
			wantNames:  []string{"left", "right"},
		},
		{
			name:       "no placeholders",
			format:     "literal text",
			wantFormat: "literal text",
			wantNames:  []string{},
		},
		{
			name:       "long placeholder list",
			format:     "%(p0)s %(p1)s %(p2)s %(p3)s %(p4)s %(p5)s %(p6)s %(p7)s %(p8)s %(p9)s",
			wantFormat: "%s %s %s %s %s %s %s %s %s %s",
			wantNames:  []string{"p0", "p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"},
		},
		{
			name:       "malformed placeholders",
			format:     "%(bad-name)s %(unterminated",
			wantFormat: "%(bad-name)s %(unterminated",
			wantNames:  []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotFormat, gotNames := reformatSprintf(test.format)
			if gotFormat != test.wantFormat {
				t.Errorf("rewritten format = %q, want %q", gotFormat, test.wantFormat)
			}
			if !reflect.DeepEqual(gotNames, test.wantNames) {
				t.Errorf("ordered names = %v, want %v", gotNames, test.wantNames)
			}
		})
	}
}

func TestParseSprintfPreservesArgumentOrder(t *testing.T) {
	format := "%(first)s%(second)d%(first)s"
	params := map[string]any{
		"first":  "one",
		"second": 2,
	}

	gotFormat, gotParams := parseSprintf(format, params)
	if gotFormat != "%s%d%s" {
		t.Errorf("rewritten format = %q, want %q", gotFormat, "%s%d%s")
	}
	wantParams := []any{"one", 2, "one"}
	if !reflect.DeepEqual(gotParams, wantParams) {
		t.Errorf("ordered params = %v, want %v", gotParams, wantParams)
	}

	_, gotParams = parseSprintf("%(bad-name)s", params)
	if gotParams != nil {
		t.Errorf("malformed format should have no bound params, got %v", gotParams)
	}
}
