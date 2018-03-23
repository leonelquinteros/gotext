/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package plurals

import (
	"encoding/json"
	"os"
	"testing"
)

type fixture struct {
	PluralForm string
	Fixture    []int
}

func TestCompiler(t *testing.T) {
	f, err := os.Open("testdata/pluralforms.json")
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(f)
	var fixtures []fixture
	err = dec.Decode(&fixtures)
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range fixtures {
		expr, err := Compile(data.PluralForm)
		if err != nil {
			t.Errorf("'%s' triggered error: %s", data.PluralForm, err)
		} else if expr == nil {
			t.Logf("'%s' compiled to nil", data.PluralForm)
			t.Fail()
		} else {
			for n, e := range data.Fixture {
				i := expr.Eval(uint32(n))
				if i != e {
					t.Logf("'%s' with n = %d, expected %d, got %d, compiled to %s", data.PluralForm, n, e, i, expr)
					t.Fail()
				}
				if i == -1 {
					break
				}
			}
		}
	}
}
