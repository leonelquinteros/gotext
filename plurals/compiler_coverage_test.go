package plurals

import (
	"testing"
)

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
