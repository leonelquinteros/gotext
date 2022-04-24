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

	lines := strings.Split(str, "\n")

	// Single line msgid / entry starts with msgid "text"
	if len(lines) == 1 {
		// Single line
		return fmt.Sprintf(`"%s"`, lines[0])
	} else if len(lines) == 2 && lines[1] == "" {
		// Single line with newline at end
		return fmt.Sprintf(`"%s\n"`, lines[0])
	}

	// Multiline msgid // entry starts with msgid ""\n"text"
	lastIdx := len(lines) - 1
	result := fmt.Sprintf("\"\"\n\"%s\\n\"", lines[0])
	for _, l := range lines[1:lastIdx] {
		result += fmt.Sprintf("\n\"%s\\n\"", l)
	}

	// If the last element is empty, the previous element ended with a newline.
	// Then the text to translate should also end with a newline and we can ignore the last entry.
	if lines[lastIdx] != "" {
		result += fmt.Sprintf("\n\"%s\"", lines[lastIdx])
	}

	return result
}
