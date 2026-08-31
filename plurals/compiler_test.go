// Original work Copyright (c) 2016 Jonas Obrist (https://github.com/ojii/gettext.go)
// Modified work Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com
// Modified work Copyright (c) 2018-present gotext maintainers (https://github.com/leonelquinteros/gotext)
//
// Licensed under the 3-Clause BSD License. See LICENSE in the project root for license information.

package plurals

import (
	"encoding/json"
	"fmt"
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
	// Malformed false branch
	expr, err = Compile("n == 1 ? 0 :")
	if err == nil {
		t.Error("Expected error for malformed false branch")
	}
	if expr != nil {
		t.Error("Expected nil expression for malformed false branch")
	}

	// Unexpected EOF in logic test
	_, err = Compile("n >")
	if err == nil {
		t.Error("Expected error for unexpected EOF")
	}
	// Zero modulus
	for _, expression := range []string{"n % 0", "n % 0 == 0"} {
		expr, err = Compile(expression)
		if err == nil {
			t.Errorf("Expected error for zero modulus expression %q", expression)
		}
		if expr != nil {
			t.Errorf("Expected nil expression for zero modulus expression %q", expression)
		}
	}

	// Missing closing parenthesis - current compiler might not catch this strictly,
	// or it catches it in a way that doesn't return an error immediately from Compile.
	// Let's remove it if it doesn't fail, or just focus on what DOES fail.
}

func TestCompile_LogicCoverage(t *testing.T) {
	// Covering more branches in compileLogicTest and others
	tests := []string{
		"n >= 1",
		"n <= 1",
		"n % 10 == 1",
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
		{"n % 10", 3, 0},  // n % 10 == 0? 3 % 10 is 3, so false (0)
		{"n % 10", 10, 1}, // 10 % 10 is 0, so true (1)
		{"n % 10 == 3", 3, 1},
		{"n % 10 == 3", 13, 1},
		{"n % 10 == 3", 4, 0},
		{"3 == n % 10", 3, 1}, // Test flipped side
		// Asian (1 form)
		{"0", 1, 0},
		{"0", 10, 0},
		// Germanic/Latin (2 forms)
		{"n != 1", 1, 0},
		{"n != 1", 2, 1},
		// French (2 forms)
		{"n > 1", 1, 0},
		{"n > 1", 2, 1},
		// Celtic (5 forms)
		{"n==1 ? 0 : n==2 ? 1 : n<7 ? 2 : n<11 ? 3 : 4", 1, 0},
		{"n==1 ? 0 : n==2 ? 1 : n<7 ? 2 : n<11 ? 3 : 4", 2, 1},
		{"n==1 ? 0 : n==2 ? 1 : n<7 ? 2 : n<11 ? 3 : 4", 5, 2},
		{"n==1 ? 0 : n==2 ? 1 : n<7 ? 2 : n<11 ? 3 : 4", 10, 3},
		{"n==1 ? 0 : n==2 ? 1 : n<7 ? 2 : n<11 ? 3 : 4", 11, 4},
		// Slavic-like (nested conditions with AND/OR)
		{"n%10==1 && n%100!=11 ? 0 : 1", 1, 0},
		{"n%10==1 && n%100!=11 ? 0 : 1", 11, 1},
		{"n%10==1 && n%100!=11 ? 0 : 1", 21, 0},
		// Baltic (3 forms)
		{"n%10==1 && n%100!=11 ? 0 : n != 0 ? 1 : 2", 1, 0},
		{"n%10==1 && n%100!=11 ? 0 : n != 0 ? 1 : 2", 2, 1},
		{"n%10==1 && n%100!=11 ? 0 : n != 0 ? 1 : 2", 0, 2},
		// Arabic (6 forms)
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 0, 0},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 1, 1},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 2, 2},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 3, 3},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 10, 3},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 11, 4},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 99, 4},
		{"n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 ? 4 : 5", 100, 5},
		// Slavic (3 forms) - using parentheses
		{"n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2", 1, 0},
		{"n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2", 2, 1},
		{"n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2", 5, 2},
		{"n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2", 11, 2},
		{"n%10==1 && n%100!=11 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2", 22, 1},
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

func TestCompile_ParenthesesAndAssociativity(t *testing.T) {
	tests := []struct {
		name string
		expr string
		n    uint32
		want int
	}{
		{name: "independent parenthesized operands", expr: "(n==1)||(n==2)", n: 1, want: 1},
		{name: "independent parenthesized operands second", expr: "(n==1)||(n==2)", n: 2, want: 1},
		{name: "independent parenthesized operands neither", expr: "(n==1)||(n==2)", n: 3, want: 0},
		{name: "redundant nesting", expr: "((((n==1))))", n: 1, want: 1},
		{name: "redundant nesting false", expr: "((((n==1))))", n: 2, want: 0},
		{name: "nested ternary in true arm", expr: "n==1 ? n==2 ? 10 : 11 : 12", n: 1, want: 11},
		{name: "nested ternary outer false", expr: "n==1 ? n==2 ? 10 : 11 : 12", n: 2, want: 12},
		{name: "nested ternary nested true arm", expr: "n==1 ? n==2 ? 10 : n==3 ? 11 : 12 : 13", n: 1, want: 12},
		{name: "established nested parenthesized test zero", expr: "n==1?0:(n==0||(n%100>0&&n%100<20))?1:2", n: 0, want: 1},
		{name: "established nested parenthesized test one", expr: "n==1?0:(n==0||(n%100>0&&n%100<20))?1:2", n: 1, want: 0},
		{name: "established nested parenthesized test teen", expr: "n==1?0:(n==0||(n%100>0&&n%100<20))?1:2", n: 19, want: 1},
		{name: "established nested parenthesized test outside range", expr: "n==1?0:(n==0||(n%100>0&&n%100<20))?1:2", n: 20, want: 2},
	}
	for _, tt := range tests {
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

func TestCompile_WhitespaceAndPrecedence(t *testing.T) {
	for _, input := range []string{" 0 ", "\t0\n", "\u20030\u2003"} {
		t.Run(fmt.Sprintf("zero_%q", input), func(t *testing.T) {
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

	tests := []struct {
		name string
		expr string
		n    uint32
		want int
	}{
		{name: "and binds tighter than or", expr: "n==1 || n==2 && n==3 ? 1 : 0", n: 1, want: 1},
		{name: "and binds tighter than or false", expr: "n==1 || n==2 && n==3 ? 1 : 0", n: 2, want: 0},
		{name: "and binds tighter than or right", expr: "n==1 && n==2 || n==3 ? 1 : 0", n: 3, want: 1},
		{name: "and binds tighter than or right false", expr: "n==1 && n==2 || n==3 ? 1 : 0", n: 1, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", tt.expr, err)
			}
			if got := expr.Eval(tt.n); got != tt.want {
				t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, tt.want)
			}
		})
	}
}

func TestCompile_ReversedRelationalAndUint32Boundaries(t *testing.T) {
	tests := []struct {
		name string
		expr string
		n    uint32
		want int
	}{
		{name: "reversed greater", expr: "1 > n", n: 0, want: 1},
		{name: "reversed greater false", expr: "1 > n", n: 1, want: 0},
		{name: "reversed greater equal", expr: "1 >= n", n: 1, want: 1},
		{name: "reversed less", expr: "1 < n", n: 2, want: 1},
		{name: "reversed less false", expr: "1 < n", n: 1, want: 0},
		{name: "reversed less equal", expr: "1 <= n", n: 1, want: 1},
		{name: "maximum equality", expr: "n == 4294967295", n: ^uint32(0), want: 1},
		{name: "maximum modulo", expr: "n % 4294967295 == 0", n: ^uint32(0), want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", tt.expr, err)
			}
			if got := expr.Eval(tt.n); got != tt.want {
				t.Errorf("Eval(%q, %d) = %d, want %d", tt.expr, tt.n, got, tt.want)
			}
		})
	}
}

func TestCompile_ResultLiteralIntRange(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	maxIntLiteral := strconv.FormatInt(int64(maxInt), 10)
	expr, err := Compile("n==1?" + maxIntLiteral + ":0")
	if err != nil {
		t.Fatalf("Compile(max-int result literal) failed: %v", err)
	}
	if expr == nil {
		t.Fatal("Compile(max-int result literal) returned a nil expression")
	}
	if got := expr.Eval(1); got != maxInt {
		t.Errorf("Eval(max-int result literal, 1) = %d, want %d", got, maxInt)
	}

	overflowLiteral := strconv.FormatUint(uint64(maxInt)+1, 10)
	expr, err = Compile("n==1?" + overflowLiteral + ":0")
	if err == nil {
		t.Error("Compile(result literal above max int) returned no error")
	}
	if expr != nil {
		t.Error("Compile(result literal above max int) returned a non-nil expression")
	}
}

func TestCompile_DeeplyNestedParentheses(t *testing.T) {
	const depth = 32 * 1024
	input := strings.Repeat("(", depth) + "n==1" + strings.Repeat(")", depth)

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

func TestCompile_MalformedExpressionsReturnNil(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	overflowLiteral := strconv.FormatUint(uint64(maxInt)+1, 10)
	expressions := []string{
		"",
		"?1:0",
		"n==1?:0",
		"n==1?0:",
		"n==1?",
		"n==1:0",
		"n==1??0:1",
		"n==1?0::1",
		"n==1?0:1:2",
		"n==1&&",
		"n==1||",
		"n==1&&&&n==2",
		"n==1||||n==2",
		"n==1&&||n==2",
		"n+@",
		"n&1",
		"n|1",
		"n==1)",
		"(n==1",
		"((n==1)",
		"n==4294967296",
		"n%4294967296==0",
		"n==1?" + overflowLiteral + ":0",
	}
	for _, input := range expressions {
		t.Run(fmt.Sprintf("malformed_%q", input), func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("Compile(%q) panicked: %v", input, recovered)
				}
			}()
			expr, err := Compile(input)
			if err == nil {
				t.Fatalf("Compile(%q) returned no error", input)
			}
			if expr != nil {
				t.Fatalf("Compile(%q) returned non-nil expression on error", input)
			}
		})
	}
}

type panickingTest struct{}

func (panickingTest) test(uint32) bool {
	panic("inactive test evaluated")
}

type panickingExpression struct{}

func (panickingExpression) Eval(uint32) int {
	panic("inactive expression evaluated")
}

func TestPluralShortCircuiting(t *testing.T) {
	if got := (and{left: equal{value: 0}, right: panickingTest{}}).test(1); got {
		t.Error("and evaluated an inactive right-hand test")
	}
	if got := (or{left: equal{value: 1}, right: panickingTest{}}).test(1); !got {
		t.Error("or did not preserve a true left-hand test")
	}
	if got := (ternary{
		test:      equal{value: 0},
		trueExpr:  panickingExpression{},
		falseExpr: constValue{value: 2},
	}).Eval(1); got != 2 {
		t.Errorf("ternary selected %d, want 2", got)
	}
	if got := (ternary{
		test:      equal{value: 1},
		trueExpr:  constValue{value: 3},
		falseExpr: panickingExpression{},
	}).Eval(1); got != 3 {
		t.Errorf("ternary selected %d, want 3", got)
	}
}

func TestPluralZeroValues(t *testing.T) {
	if got := (ternary{}).Eval(0); got != -1 {
		t.Errorf("zero ternary Eval() = %d, want -1", got)
	}
	if got := (and{}).test(0); got {
		t.Error("zero and test returned true")
	}
	if got := (or{}).test(0); got {
		t.Error("zero or test returned true")
	}
	if got := (pipe{}).test(0); got {
		t.Error("zero pipe test returned true")
	}
	if got := (mod{}).calc(7); got != 0 {
		t.Errorf("zero mod calc() = %d, want 0", got)
	}
}

func FuzzCompileEvalNoPanic(f *testing.F) {
	for _, seed := range []string{
		"0",
		" n == 1 ? 0 : 1 ",
		"(n==1)||(n==2)",
		"n==1?n==2?1:n==3?2:3:4",
		"n%4294967295==0",
		"",
		"n==1&&",
		"(n==1",
		"n==4294967296",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 64<<10 {
			return
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("Compile(%q) panicked: %v", input, recovered)
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
		source := fmt.Sprintf("n%%%d==%d", divisor, value)
		expr, err := Compile(source)
		if err != nil {
			t.Fatalf("Compile(%q) failed: %v", source, err)
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
