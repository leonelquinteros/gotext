package parser

import (
	"go/parser"
	"testing"
)

func TestPrepareString(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "Quotation marks are added",
			raw:  `Extracted string`,
			want: `"Extracted string"`,
		},
		{
			name: "Intentional multiple quotation marks are preserved",
			raw:  `"Extracted string"`,
			want: `""Extracted string""`,
		},
		{
			name: "Intentional backquotes are preserved",
			raw:  "`Extracted string`",
			want: "\"`Extracted string`\"",
		},
		{
			name: "Multiline text are formatted correctly",
			raw:  "This is an multiline\nstring",
			want: `""
"This is an multiline\n"
"string"`,
		},
		{
			name: "backquoted newline is converted to newline",
			raw: `This is an multiline
string`,
			want: `""
"This is an multiline\n"
"string"`,
		},
		{
			name: "Single line with a newline at the end remains a single line",
			raw:  "Single line with newline\n",
			want: `"Single line with newline\n"`,
		},
		{
			name: "Last newline does not start a new line",
			raw:  "Multiline\nwith\nnewlines\n",
			want: `""
"Multiline\n"
"with\n"
"newlines\n"`,
		},
		{
			name: "Empty string is ignored",
			raw:  "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareString(tt.raw); got != tt.want {
				t.Errorf("PrepareString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractStringLiteral(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantStr   string
		wantFound bool
	}{
		{
			name:      "String extracted",
			code:      `"Extracted string"`,
			wantStr:   `"Extracted string"`,
			wantFound: true,
		},
		{
			name:      "Even addition is merged",
			code:      `"Extracted " + "string"`,
			wantStr:   `"Extracted string"`,
			wantFound: true,
		},
		{
			name:      "Odd addition is merged",
			code:      `"Extracted " + "string" + " is combined"`,
			wantStr:   `"Extracted string is combined"`,
			wantFound: true,
		},
		{
			name:      "Backquotes are replaced",
			code:      "`Extracted string`",
			wantStr:   "\"Extracted string\"",
			wantFound: true,
		},
		{
			name: "Multiline text with backquotes are formatted correctly",
			code: "`This is an multiline\nstring`",
			wantStr: `""
"This is an multiline\n"
"string"`,
			wantFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.code)
			if err != nil {
				t.Errorf("Expression %s could not be parsed: %v", tt.code, expr)
			}
			extractedStr, found := ExtractStringLiteral(expr)
			if extractedStr != tt.wantStr {
				t.Errorf("ExtractStringLiteral() string = %v, want %v", extractedStr, tt.wantStr)
			}
			if found != tt.wantFound {
				t.Errorf("ExtractStringLiteral() got1 = %v, want %v", found, tt.wantFound)
			}
		})
	}

	t.Run("Nil is ignored", func(t *testing.T) {
		extractedStr, found := ExtractStringLiteral(nil)
		if extractedStr != "" {
			t.Errorf("ExtractStringLiteral() string = %v, want %v", extractedStr, "")
		}
		if found != false {
			t.Errorf("ExtractStringLiteral() got1 = %v, want %v", found, false)
		}
	})
}
