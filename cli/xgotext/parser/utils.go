package parser

import (
	"go/ast"
	"go/token"
	"strconv"
)

func PrepareString(rawString string) string {
	// Remove backquotes and add quotes
	unquoteString, err := strconv.Unquote(rawString)
	if err != nil {
		return rawString
	}

	return strconv.Quote(unquoteString)
}

// ExtractStringLiteral checks if an expression is a string and returns it.
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

	return strconv.Quote(result), true
}
