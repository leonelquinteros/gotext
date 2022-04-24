package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// ExtractStringLiteral checks if an expression represents a string and returns it correctly formatted.
func ExtractStringLiteral(expr ast.Expr) (string, bool) {
	stack := []ast.Expr{expr}
	var b strings.Builder

	for len(stack) != 0 {
		n := len(stack) - 1
		elem := stack[n]
		stack = stack[:n]

		switch v := elem.(type) {
		//  Simple string with quotes or backqoutes
		case *ast.BasicLit:
			if v.Kind != token.STRING {
				return "", false
			}

			if unqouted, err := strconv.Unquote(v.Value); err != nil {
				b.WriteString(v.Value)
			} else {
				b.WriteString(unqouted)
			}
		// Concatenation of several string literals
		case *ast.BinaryExpr:
			if v.Op != token.ADD {
				return "", false
			}
			stack = append(stack, v.Y, v.X)
		default:
			return "", false
		}
	}

	return prepareString(b.String()), true
}

func prepareString(str string) string {
	if str == "" {
		return ""
	}

	// Entry starts with msgid "text"
	lines := strings.Split(str, "\n")
	if len(lines) == 1 {
		return fmt.Sprintf("\"%s\"", lines[0])
	}

	lastIdx := len(lines) - 1
	result := "\"\"\n" // Entry starts with msgid ""\n"text"
	for _, line := range lines[:lastIdx] {
		result += fmt.Sprintf("\"%s\\n\"\n", line)
	}
	result += fmt.Sprintf("\"%s\"", lines[lastIdx])

	return result
}
