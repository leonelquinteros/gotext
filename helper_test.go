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
	params := map[string]interface{}{
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
	params := map[string]interface{}{
		"sister":  "Susan",
		"brother": "Louis",
	}

	NPrintf(pat, params)

}

func TestSprintfFloatsWithPrecision(t *testing.T) {
	pat := "%(float)f / %(floatprecision).1f / %(long)g / %(longprecision).3g"
	params := map[string]interface{}{
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
