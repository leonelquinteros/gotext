package parser

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// ExtractStringLiteral checks if an expression represents a string and returns it correctly formatted.
func ExtractStringLiteral(expr ast.Expr) (string, bool) {
	stack := []ast.Expr{expr}
	result := ""

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
				result = v.Value + result
			} else {
				result = unqouted + result
			}
		// Concatenation of several string literals
		case *ast.BinaryExpr:
			if v.Op != token.ADD {
				return "", false
			}
			stack = append(stack, v.X, v.Y)
		default:
			return "", false
		}
	}

	return prepareString(result), true
}

func prepareString(rawString string) string {
	if strings.HasPrefix(rawString, `"`) && strings.HasSuffix(rawString, `"`) {
		return rawString
	}

	// Remove backquotes and add quotes
	unquoteString, err := strconv.Unquote(rawString)
	if err != nil {
		return strconv.Quote(rawString)
	}

	return strconv.Quote(unquoteString)
}
