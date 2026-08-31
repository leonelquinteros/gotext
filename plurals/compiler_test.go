// Original work Copyright (c) 2016 Jonas Obrist (https://github.com/ojii/gettext.go)
// Modified work Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com
// Modified work Copyright (c) 2018-present gotext maintainers (https://github.com/leonelquinteros/gotext)
//
// Licensed under the 3-Clause BSD License. See LICENSE in the project root for license information.

package plurals

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
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
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close plural forms fixture: %v", err)
		}
	})

	var fixtures []fixture
	if err := json.NewDecoder(f).Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	for _, data := range fixtures {
		data := data
		t.Run(data.PluralForm, func(t *testing.T) {
			expr, err := Compile(data.PluralForm)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", data.PluralForm, err)
			}
			if expr == nil {
				t.Fatalf("Compile(%q) returned a nil expression", data.PluralForm)
			}
			for n, want := range data.Fixture {
				if got := expr.Eval(uint32(n)); got != want {
					t.Errorf("Eval(%q, %d) = %d, want %d", data.PluralForm, n, got, want)
				}
			}
		})
	}
}

func TestCompileEvalGNUExpressions(t *testing.T) {
	tests := []struct {
		name string
		expr string
		n    uint32
		want int
	}{
		{name: "implicit equality true", expr: "n", n: 1, want: 1},
		{name: "implicit equality false", expr: "n", n: 2, want: 0},
		{name: "equality true", expr: "n == 1", n: 1, want: 1},
		{name: "equality false", expr: "n == 1", n: 2, want: 0},
		{name: "inequality true", expr: "n != 1", n: 2, want: 1},
		{name: "inequality false", expr: "n != 1", n: 1, want: 0},
		{name: "greater true", expr: "n > 1", n: 2, want: 1},
		{name: "greater false", expr: "n > 1", n: 1, want: 0},
		{name: "greater equal boundary", expr: "n >= 1", n: 1, want: 1},
		{name: "greater equal false", expr: "n >= 1", n: 0, want: 0},
		{name: "less true", expr: "n < 2", n: 1, want: 1},
		{name: "less false", expr: "n < 2", n: 2, want: 0},
		{name: "less equal boundary", expr: "n <= 1", n: 1, want: 1},
		{name: "less equal false", expr: "n <= 1", n: 2, want: 0},
		{name: "modulo default true", expr: "n % 10", n: 10, want: 1},
		{name: "modulo default false", expr: "n % 10", n: 3, want: 0},
		{name: "modulo equality", expr: "n % 10 == 3", n: 13, want: 1},
		{name: "modulo equality false", expr: "n % 10 == 3", n: 14, want: 0},
		{name: "reversed modulo equality", expr: "3 == n % 10", n: 13, want: 1},
		{name: "reversed modulo inequality", expr: "3 != n % 10", n: 13, want: 0},
		{name: "simple ternary true", expr: "n == 1 ? 7 : 9", n: 1, want: 7},
		{name: "simple ternary false", expr: "n == 1 ? 7 : 9", n: 2, want: 9},
		{name: "nested ternary inner true", expr: "n < 3 ? n == 2 ? 10 : 11 : 12", n: 2, want: 10},
		{name: "nested ternary inner false", expr: "n < 3 ? n == 2 ? 10 : 11 : 12", n: 1, want: 11},
		{name: "nested ternary outer false", expr: "n < 3 ? n == 2 ? 10 : 11 : 12", n: 3, want: 12},
		{name: "nested false arm", expr: "n == 0 ? 0 : n == 1 ? 1 : 2", n: 2, want: 2},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", tt.expr, err)
			}
			if expr == nil {
				t.Fatalf("Compile(%q) returned a nil expression", tt.expr)
			}
			if got := expr.Eval(tt.n); got != tt.want {
				t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, tt.want)
			}
		})
	}
}

func TestCompileEvalPrecedenceAndParentheses(t *testing.T) {
	tests := []struct {
		name string
		expr string
		n    uint32
		want int
	}{
		{
			name: "and binds tighter than or, left branch",
			expr: "n == 1 || n == 2 && n != 1 ? 1 : 0",
			n:    1,
			want: 1,
		},
		{
			name: "and binds tighter than or, right branch",
			expr: "n == 1 || n == 2 && n != 1 ? 1 : 0",
			n:    2,
			want: 1,
		},
		{
			name: "and binds tighter than or, neither branch",
			expr: "n == 1 || n == 2 && n != 1 ? 1 : 0",
			n:    3,
			want: 0,
		},
		{
			name: "or follows and",
			expr: "n == 1 && n != 2 || n == 3 ? 1 : 0",
			n:    3,
			want: 1,
		},
		{
			name: "parentheses override and-or precedence",
			expr: "(n == 1 || n == 2) && n != 1 ? 1 : 0",
			n:    1,
			want: 0,
		},
		{
			name: "parentheses preserve selected branch",
			expr: "(n == 1 || n == 2) && n != 1 ? 1 : 0",
			n:    2,
			want: 1,
		},
		{
			name: "nested parenthesized plural rule zero",
			expr: "n == 1 ? 0 : (n == 0 || (n % 100 > 0 && n % 100 < 20)) ? 1 : 2",
			n:    0,
			want: 1,
		},
		{
			name: "nested parenthesized plural rule one",
			expr: "n == 1 ? 0 : (n == 0 || (n % 100 > 0 && n % 100 < 20)) ? 1 : 2",
			n:    1,
			want: 0,
		},
		{
			name: "nested parenthesized plural rule teen",
			expr: "n == 1 ? 0 : (n == 0 || (n % 100 > 0 && n % 100 < 20)) ? 1 : 2",
			n:    19,
			want: 1,
		},
		{
			name: "nested parenthesized plural rule outside range",
			expr: "n == 1 ? 0 : (n == 0 || (n % 100 > 0 && n % 100 < 20)) ? 1 : 2",
			n:    20,
			want: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", tt.expr, err)
			}
			if expr == nil {
				t.Fatalf("Compile(%q) returned a nil expression", tt.expr)
			}
			if got := expr.Eval(tt.n); got != tt.want {
				t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, tt.want)
			}
		})
	}
}

func TestCompileEvalLogicalShortCircuitCases(t *testing.T) {
	// GNU plural expressions are side-effect free, so the source-level
	// short-circuit contract is checked against native Go boolean semantics.
	tests := []struct {
		name string
		expr string
		n    uint32
		want func(uint32) bool
	}{
		{
			name: "false and",
			expr: "n == 0 && n % 10 == 0 ? 1 : 0",
			n:    1,
			want: func(n uint32) bool { return n == 0 && n%10 == 0 },
		},
		{
			name: "true and with false right side",
			expr: "n == 1 && n % 10 == 0 ? 1 : 0",
			n:    1,
			want: func(n uint32) bool { return n == 1 && n%10 == 0 },
		},
		{
			name: "true or",
			expr: "n == 1 || n % 10 == 0 ? 1 : 0",
			n:    1,
			want: func(n uint32) bool { return n == 1 || n%10 == 0 },
		},
		{
			name: "false or with true right side",
			expr: "n == 1 || n % 10 == 0 ? 1 : 0",
			n:    10,
			want: func(n uint32) bool { return n == 1 || n%10 == 0 },
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", tt.expr, err)
			}
			if expr == nil {
				t.Fatalf("Compile(%q) returned a nil expression", tt.expr)
			}
			want := 0
			if tt.want(tt.n) {
				want = 1
			}
			if got := expr.Eval(tt.n); got != want {
				t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, want)
			}
		})
	}
}

func TestCompileEvalWhitespaceAndUint32Boundaries(t *testing.T) {
	for _, input := range []string{" 0 ", "\t0\n", "\u20030\u2003"} {
		input := input
		t.Run("constant whitespace", func(t *testing.T) {
			expr, err := Compile(input)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", input, err)
			}
			if expr == nil {
				t.Fatalf("Compile(%q) returned a nil expression", input)
			}
			if got := expr.Eval(^uint32(0)); got != 0 {
				t.Errorf("Eval(%q, max uint32) = %d, want 0", input, got)
			}
		})
	}

	const maxUint32 = ^uint32(0)
	maxLiteral := strconv.FormatUint(uint64(maxUint32), 10)
	tests := []struct {
		name string
		expr string
		n    uint32
		want int
	}{
		{name: "reversed greater true", expr: "1 > n", n: 0, want: 1},
		{name: "reversed greater false", expr: "1 > n", n: 1, want: 0},
		{name: "reversed greater equal", expr: "1 >= n", n: 1, want: 1},
		{name: "reversed less true", expr: "1 < n", n: 2, want: 1},
		{name: "reversed less false", expr: "1 < n", n: 1, want: 0},
		{name: "reversed less equal", expr: "1 <= n", n: 1, want: 1},
		{name: "maximum equality", expr: "n == " + maxLiteral, n: maxUint32, want: 1},
		{name: "maximum equality false", expr: "n == " + maxLiteral, n: maxUint32 - 1, want: 0},
		{name: "maximum modulo remainder", expr: "n % " + maxLiteral + " == 0", n: maxUint32, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", tt.expr, err)
			}
			if expr == nil {
				t.Fatalf("Compile(%q) returned a nil expression", tt.expr)
			}
			if got := expr.Eval(tt.n); got != tt.want {
				t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, tt.want)
			}
		})
	}
}

func TestCompileResultLiteralIntRange(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	maxIntLiteral := strconv.FormatInt(int64(maxInt), 10)
	expr, err := Compile("n == 1 ? " + maxIntLiteral + " : 0")
	if err != nil {
		t.Fatalf("Compile(max-int result literal) failed: %v", err)
	}
	if expr == nil {
		t.Fatal("Compile(max-int result literal) returned a nil expression")
	}
	if got := expr.Eval(1); got != maxInt {
		t.Errorf("Eval(max-int result literal, 1) = %d, want %d", got, maxInt)
	}
	if got := expr.Eval(2); got != 0 {
		t.Errorf("Eval(max-int result literal, 2) = %d, want 0", got)
	}

	overflowLiteral := strconv.FormatUint(uint64(maxInt)+1, 10)
	expr, err = Compile("n == 1 ? " + overflowLiteral + " : 0")
	if err == nil {
		t.Error("Compile(result literal above max int) returned no error")
	}
	if expr != nil {
		t.Error("Compile(result literal above max int) returned a non-nil expression")
	}
}

func TestCompileDeeplyNestedParentheses(t *testing.T) {
	const depth = 32 * 1024
	input := strings.Repeat("(", depth) + "n == 1" + strings.Repeat(")", depth)

	expr, err := Compile(input)
	if err != nil {
		t.Fatalf("Compile(%d nested parentheses) failed: %v", depth, err)
	}
	if expr == nil {
		t.Fatalf("Compile(%d nested parentheses) returned a nil expression", depth)
	}
	if got := expr.Eval(1); got != 1 {
		t.Errorf("Eval(%d nested parentheses, 1) = %d, want 1", depth, got)
	}
	if got := expr.Eval(2); got != 0 {
		t.Errorf("Eval(%d nested parentheses, 2) = %d, want 0", depth, got)
	}
}

func TestCompileMalformedExpressionsReturnNil(t *testing.T) {
	maxUint32 := ^uint32(0)
	aboveUint32 := strconv.FormatUint(uint64(maxUint32)+1, 10)
	maxInt := int(^uint(0) >> 1)
	aboveMaxInt := strconv.FormatUint(uint64(maxInt)+1, 10)
	tests := []struct {
		name string
		expr string
	}{
		{name: "empty expression", expr: ""},
		{name: "missing ternary condition", expr: "?1:0"},
		{name: "missing true branch", expr: "n == 1 ? : 0"},
		{name: "missing false branch", expr: "n == 1 ? 0 :"},
		{name: "missing ternary colon", expr: "n == 1 ?"},
		{name: "colon without ternary", expr: "n == 1 : 0"},
		{name: "nested ternary missing true branch", expr: "n == 1 ? ?0:1"},
		{name: "nested ternary extra colon", expr: "n == 1 ? 0::1"},
		{name: "extra ternary branch", expr: "n == 1 ? 0 : 1 : 2"},
		{name: "missing and operand", expr: "n == 1 &&"},
		{name: "missing or operand", expr: "n == 1 ||"},
		{name: "repeated and operator", expr: "n == 1 &&&& n == 2"},
		{name: "repeated or operator", expr: "n == 1 |||| n == 2"},
		{name: "mixed logical operators", expr: "n == 1 && || n == 2"},
		{name: "invalid character", expr: "n + @"},
		{name: "single ampersand", expr: "n&1"},
		{name: "single pipe", expr: "n|1"},
		{name: "unmatched closing parenthesis", expr: "n == 1)"},
		{name: "unmatched opening parenthesis", expr: "(n == 1"},
		{name: "multiple unmatched opening parentheses", expr: "((n == 1)"},
		{name: "uint32 literal overflow", expr: "n == " + aboveUint32},
		{name: "modulus literal overflow", expr: "n % " + aboveUint32 + " == 0"},
		{name: "zero modulus", expr: "n % 0"},
		{name: "zero modulus comparison", expr: "n % 0 == 0"},
		{name: "missing modulus right operand", expr: "n %"},
		{name: "comparison without n", expr: "1 == 2"},
		{name: "comparison of two n values", expr: "n == n"},
		{name: "result literal overflow", expr: "n == 1 ? " + aboveMaxInt + " : 0"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("Compile(%q) panicked: %v", tt.expr, recovered)
				}
			}()
			expr, err := Compile(tt.expr)
			if err == nil {
				t.Fatalf("Compile(%q) returned no error", tt.expr)
			}
			if expr != nil {
				t.Fatalf("Compile(%q) returned non-nil expression on error", tt.expr)
			}
		})
	}
}

func FuzzCompileEvalNoPanic(f *testing.F) {
	for _, seed := range []string{
		"0",
		" n == 1 ? 0 : 1 ",
		"(n == 1) || (n == 2)",
		"n == 1 ? n == 2 ? 1 : n == 3 ? 2 : 3 : 4",
		"n % 4294967295 == 0",
		"",
		"n == 1 &&",
		"(n == 1",
		"n == 4294967296",
	} {
		f.Add(seed)
	}

	const maxInputBytes = 64 << 10
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > maxInputBytes {
			return
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("Compile/Eval(%q) panicked: %v", input, recovered)
			}
		}()

		expr, err := Compile(input)
		if err != nil {
			if expr != nil {
				t.Fatalf("Compile(%q) returned non-nil expression on error", input)
			}
			return
		}
		if expr == nil {
			t.Fatalf("Compile(%q) returned a nil expression without an error", input)
		}
		for _, n := range []uint32{0, 1, 2, ^uint32(0)} {
			_ = expr.Eval(n)
		}
	})
}

func FuzzModuloComparisonMatchesUint32(f *testing.F) {
	for _, seed := range [][3]uint32{
		{0, 1, 0},
		{1, 2, 1},
		{^uint32(0), ^uint32(0), 0},
		{^uint32(0), 3, 2},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}

	f.Fuzz(func(t *testing.T, n, divisor, value uint32) {
		if divisor == 0 {
			return
		}

		// Keep source construction and the oracle in uint32 space. This is
		// independent of the host int width on both 32-bit and 64-bit Go.
		source := "n%" + strconv.FormatUint(uint64(divisor), 10) +
			"==" + strconv.FormatUint(uint64(value), 10)
		expr, err := Compile(source)
		if err != nil {
			t.Fatalf("Compile(%q) failed: %v", source, err)
		}
		if expr == nil {
			t.Fatalf("Compile(%q) returned a nil expression", source)
		}

		want := 0
		if n%divisor == value {
			want = 1
		}
		if got := expr.Eval(n); got != want {
			t.Errorf("Eval(%q, %d) = %d, want %d", source, n, got, want)
		}
	})
}
