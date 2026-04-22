package parser

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestGetterDef_MaxArgIndex(t *testing.T) {
	tests := []struct {
		def  GetterDef
		want int
	}{
		{GetterDef{0, -1, -1, -1}, 0},
		{GetterDef{0, 1, -1, -1}, 1},
		{GetterDef{1, -1, -1, 0}, 1},
		{GetterDef{1, 2, 4, 0}, 4},
	}
	for _, tt := range tests {
		if got := tt.def.MaxArgIndex(); got != tt.want {
			t.Errorf("GetterDef.MaxArgIndex() = %v, want %v", got, tt.want)
		}
	}
}

func TestGoFile_ParseGetter(t *testing.T) {
	data := &DomainMap{}
	g := &GoFile{
		Data: data,
	}

	def := gotextGetter["Get"]
	args := []*ast.BasicLit{
		{Kind: token.STRING, Value: "\"msgid\""},
	}
	g.ParseGetter(def, args, "file.go:10")

	if len(data.Domains["default"].Translations) != 1 {
		t.Error("ParseGetter failed for simple Get")
	}

	// Test plural
	defN := gotextGetter["GetN"]
	argsN := []*ast.BasicLit{
		{Kind: token.STRING, Value: "\"singular\""},
		{Kind: token.STRING, Value: "\"plural\""},
	}
	g.ParseGetter(defN, argsN, "file.go:20")
	if data.Domains["default"].Translations["singular"].MsgIDPlural != "plural" {
		t.Error("ParseGetter failed for GetN")
	}

	// Test Domain
	defD := gotextGetter["GetD"]
	argsD := []*ast.BasicLit{
		{Kind: token.STRING, Value: "\"domain1\""},
		{Kind: token.STRING, Value: "\"msgid_d\""},
	}
	g.ParseGetter(defD, argsD, "file.go:30")
	if _, ok := data.Domains["domain1"]; !ok {
		t.Error("ParseGetter failed for GetD domain creation")
	}
}

func TestGoFile_ParseGetter_Errors(t *testing.T) {
	data := &DomainMap{}
	g := &GoFile{
		Data: data,
	}

	// Not enough args for GetN (needs 2: ID and Plural, which are index 0 and 1. MaxArgIndex is 1)
	defN := gotextGetter["GetN"]
	args := []*ast.BasicLit{
		{Kind: token.STRING, Value: "\"singular\""},
	} // only 1 arg, index 0. len(args) == 1. 1 <= 1 is true. Should return.
	g.ParseGetter(defN, args, "file.go:10")
	if len(data.Domains) != 0 {
		t.Error("ParseGetter should have failed for not enough args (len 1 for GetN)")
	}

	// Not enough args for GetD (needs 2: Domain and ID, index 0 and 1. MaxArgIndex is 1)
	defD := gotextGetter["GetD"]
	argsD := []*ast.BasicLit{
		{Kind: token.STRING, Value: "\"domain1\""},
	}
	g.ParseGetter(defD, argsD, "file.go:15")
	if len(data.Domains) != 0 {
		t.Error("ParseGetter should have failed for not enough args (len 1 for GetD)")
	}

	// ID not a string
	defGet := gotextGetter["Get"]
	args2 := []*ast.BasicLit{
		{Kind: token.INT, Value: "123"},
	}
	g.ParseGetter(defGet, args2, "file.go:20")
}

