package parser

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"golang.org/x/tools/go/packages"
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

	// Missing or unsupported arguments must not cause an index-out-of-range panic.
	g.ParseGetter(defGet, nil, "file.go:30")
	g.ParseGetter(gotextGetter["GetD"], []*ast.BasicLit{nil}, "file.go:40")
}

func TestGetLiteralFromIdent(t *testing.T) {
	// ValueSpec multi-var: const (A = "valA"; B = "valB")
	valA := &ast.BasicLit{Kind: token.STRING, Value: `"valA"`}
	valB := &ast.BasicLit{Kind: token.STRING, Value: `"valB"`}
	identA := &ast.Ident{Name: "A"}
	identB := &ast.Ident{Name: "B"}
	spec := &ast.ValueSpec{
		Names:  []*ast.Ident{identA, identB},
		Values: []ast.Expr{valA, valB},
	}
	identA.Obj = &ast.Object{Decl: spec} //nolint:staticcheck
	identB.Obj = &ast.Object{Decl: spec} //nolint:staticcheck

	litA := getLiteralFromIdent(identA)
	if litA == nil || litA.Value != `"valA"` {
		t.Errorf("expected valA, got %v", litA)
	}

	litB := getLiteralFromIdent(identB)
	if litB == nil || litB.Value != `"valB"` {
		t.Errorf("expected valB, got %v", litB)
	}

	// AssignStmt multi-var: x, y := "valX", "valY"
	valX := &ast.BasicLit{Kind: token.STRING, Value: `"valX"`}
	valY := &ast.BasicLit{Kind: token.STRING, Value: `"valY"`}
	identX := &ast.Ident{Name: "x"}
	identY := &ast.Ident{Name: "y"}
	assign := &ast.AssignStmt{
		Lhs: []ast.Expr{identX, identY},
		Rhs: []ast.Expr{valX, valY},
	}
	identX.Obj = &ast.Object{Decl: assign} //nolint:staticcheck
	identY.Obj = &ast.Object{Decl: assign} //nolint:staticcheck

	litX := getLiteralFromIdent(identX)
	if litX == nil || litX.Value != `"valX"` {
		t.Errorf("expected valX, got %v", litX)
	}

	litY := getLiteralFromIdent(identY)
	if litY == nil || litY.Value != `"valY"` {
		t.Errorf("expected valY, got %v", litY)
	}
}

type typedASTImporter struct {
	pkg *types.Package
}

func (i typedASTImporter) Import(string) (*types.Package, error) {
	return i.pkg, nil
}

func addTypedGetter(pkg *types.Package, name string, parameterTypes ...types.Type) {
	params := make([]*types.Var, len(parameterTypes))
	for idx, parameterType := range parameterTypes {
		params[idx] = types.NewVar(token.NoPos, pkg, "", parameterType)
	}

	signature := types.NewSignatureType(nil, nil, nil, types.NewTuple(params...), types.NewTuple(), false)
	pkg.Scope().Insert(types.NewFunc(token.NoPos, pkg, name, signature))
}

func newTypedGoFile(t *testing.T, source string) (*GoFile, *ast.File) {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := goparser.ParseFile(fileSet, "typed.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

	gotextTypes := types.NewPackage("github.com/leonelquinteros/gotext", "gotext")
	addTypedGetter(gotextTypes, "Get", types.Typ[types.String])
	addTypedGetter(gotextTypes, "GetNC",
		types.Typ[types.String],
		types.Typ[types.String],
		types.Typ[types.String],
		types.Typ[types.String],
	)
	gotextTypes.MarkComplete()

	typesInfo := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	_, typeErr := (&types.Config{
		Importer: typedASTImporter{pkg: gotextTypes},
	}).Check("typed", fileSet, []*ast.File{file}, typesInfo)
	if typeErr != nil {
		t.Fatalf("type-check typed source: %v", typeErr)
	}

	g := &GoFile{
		FilePath: "typed.go",
		BasePath: ".",
		Data:     &DomainMap{},
		FileSet:  fileSet,
		ImportedPackages: map[string]*packages.Package{
			"gotext": {
				PkgPath: "github.com/leonelquinteros/gotext",
			},
			"typed": {
				PkgPath:   "typed",
				Syntax:    []*ast.File{file},
				TypesInfo: typesInfo,
			},
		},
	}
	return g, file
}

func inspectTypedCalls(g *GoFile, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			g.InspectCallExpr(call)
		}
		return true
	})
}

func makeTypedASTCycle(t *testing.T, file *ast.File) {
	t.Helper()

	declarations := make(map[string]*ast.ValueSpec, 2)
	references := make(map[string]*ast.Ident, 2)
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.ValueSpec:
			for _, name := range node.Names {
				if name.Name == "cyclicA" || name.Name == "cyclicB" {
					declarations[name.Name] = node
				}
			}
		case *ast.Ident:
			spec, ok := declarations[node.Name]
			if ok && node.Obj != nil && node.Obj.Decl == spec {
				if slices.Contains(spec.Names, node) {
					return true
				}
				references[node.Name] = node
			}
		}
		return true
	})

	for _, name := range []string{"cyclicA", "cyclicB"} {
		if declarations[name] == nil || references[name] == nil {
			t.Fatalf("failed to find typed AST cycle node %s", name)
		}
	}
	declarations["cyclicA"].Values = []ast.Expr{references["cyclicB"]}
	declarations["cyclicB"].Values = []ast.Expr{references["cyclicA"]}
}

func TestGoFile_InspectCallExpr_TypedASTMutation(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

func example() {
	unrelated := "unrelated"
	unrelated = "changed unrelated"
	_ = unrelated
	before := "message before mutation"
	gotext.Get(before)

	mutated := "initial mutation"
	mutated = "message after mutation"
	gotext.Get(mutated)

	afterCall := "message before later mutation"
	gotext.Get(afterCall)
	afterCall = "message after later mutation"
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	translations := g.Data.Domains["default"].Translations
	if len(translations) != 2 {
		t.Fatalf("got %d translations, want 2", len(translations))
	}
	if _, ok := translations["message before mutation"]; !ok {
		t.Error("unmodified identifier before its call was not extracted")
	}
	if _, ok := translations["message before later mutation"]; !ok {
		t.Error("identifier mutated after its call was not extracted")
	}
	if _, ok := translations["message after mutation"]; ok {
		t.Error("identifier mutated before its call was extracted")
	}
}

func TestGoFile_InspectCallExpr_TypedASTCycleDoesNotPoisonLaterArgumentsOrCalls(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

var cyclicA string = "initial cyclic A"
var cyclicB string = "initial cyclic B"
var contextAfterCycle = "context after cycle"
var later = "later call"

func example() {
	gotext.Get(cyclicA)
	gotext.GetNC("id", "plural", cyclicB, contextAfterCycle)
	gotext.Get(later)
}`

	g, file := newTypedGoFile(t, source)
	makeTypedASTCycle(t, file)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if _, ok := domain.Translations["cyclicA"]; ok {
		t.Error("cyclic identifier was extracted")
	}
	contextTranslations := domain.ContextTranslations[`"context after cycle"`]
	translation, ok := contextTranslations["id"]
	if !ok {
		t.Fatal("later arguments were not extracted after a cyclic identifier")
	}
	if translation.Context != `"context after cycle"` {
		t.Errorf("context = %q, want %q", translation.Context, `"context after cycle"`)
	}
	if _, ok := domain.Translations["later call"]; !ok {
		t.Error("later call was not extracted after a cyclic identifier")
	}
	if len(domain.Translations) != 1 || len(contextTranslations) != 1 {
		t.Fatalf("got %d uncontexted and %d contextual translations, want 1 each",
			len(domain.Translations), len(contextTranslations))
	}
}
