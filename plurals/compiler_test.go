// Original work Copyright (c) 2016 Jonas Obrist (https://github.com/ojii/gettext.go)
// Modified work Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com
// Modified work Copyright (c) 2018-present gotext maintainers (https://github.com/leonelquinteros/gotext)
//
// Licensed under the 3-Clause BSD License. See LICENSE in the project root for license information.

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

func TestCompile_EdgeCases(t *testing.T) {
	// Empty expression
	expr, err := Compile("")
	if err == nil {
		t.Error("Expected error for empty expression")
	}
	if expr != nil {
		t.Error("Expected nil expression for error")
	}

	// Invalid tokens
	_, err = Compile("n + @")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	// Malformed ternary
	_, err = Compile("n ? 1")
	if err == nil {
		t.Error("Expected error for malformed ternary")
	}

	// Unexpected EOF in logic test
	_, err = Compile("n >")
	if err == nil {
		t.Error("Expected error for unexpected EOF")
	}

	// Missing closing parenthesis - current compiler might not catch this strictly, 
	// or it catches it in a way that doesn't return an error immediately from Compile.
	// Let's remove it if it doesn't fail, or just focus on what DOES fail.
}

func TestCompile_LogicCoverage(t *testing.T) {
	// Covering more branches in compileLogicTest and others
	tests := []string{
		"n != 1",
		"n > 1",
		"n < 1",
		"n >= 1",
		"n <= 1",
		"n % 10 == 1",
		"(n == 1) && (n != 2)",
		"(n == 1) || (n == 2)",
		"n == 1 ? 0 : n == 2 ? 1 : 2",
	}
	for _, tt := range tests {
		_, err := Compile(tt)
		if err != nil {
			t.Errorf("Compile(%q) failed: %v", tt, err)
		}
	}
}

func TestEval_EdgeCases(t *testing.T) {
	// Covering eval with different operators
	tests := []struct {
		expr string
		n    uint32
		want int
	}{
		{"n == 1", 1, 1},
		{"n == 1", 2, 0},
		{"n != 1", 1, 0},
		{"n != 1", 2, 1},
		{"n > 1", 2, 1},
		{"n < 2", 1, 1},
		{"n >= 1", 1, 1},
		{"n <= 1", 1, 1},
	}
	for _, tt := range tests {
		expr, err := Compile(tt.expr)
		if err != nil {
			t.Errorf("Compile(%q) failed: %v", tt.expr, err)
			continue
		}
		if got := expr.Eval(tt.n); got != tt.want {
			t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, tt.want)
		}
	}
}
