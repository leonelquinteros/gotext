package parser

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"go/types"
	"slices"
	"strconv"
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

func typedSignature(
	pkg *types.Package,
	receiver types.Type,
	variadic bool,
	parameterTypes ...types.Type,
) *types.Signature {
	params := make([]*types.Var, len(parameterTypes))
	for idx, parameterType := range parameterTypes {
		if variadic && idx == len(parameterTypes)-1 {
			parameterType = types.NewSlice(parameterType)
		}
		params[idx] = types.NewVar(token.NoPos, pkg, "", parameterType)
	}

	var receiverVar *types.Var
	if receiver != nil {
		receiverVar = types.NewVar(token.NoPos, pkg, "", receiver)
	}
	return types.NewSignatureType(
		receiverVar,
		nil,
		nil,
		types.NewTuple(params...),
		types.NewTuple(),
		variadic,
	)
}

func addTypedGetterWithVariadic(pkg *types.Package, name string, variadic bool, parameterTypes ...types.Type) {
	pkg.Scope().Insert(types.NewFunc(token.NoPos, pkg, name, typedSignature(pkg, nil, variadic, parameterTypes...)))
}

func addTypedMethodWithVariadic(
	pkg *types.Package,
	receiver *types.Named,
	name string,
	variadic bool,
	parameterTypes ...types.Type,
) {
	receiver.AddMethod(types.NewFunc(
		token.NoPos,
		pkg,
		name,
		typedSignature(pkg, receiver, variadic, parameterTypes...),
	))
}

func addTypedReceiver(pkg *types.Package, name string, anyType types.Type) *types.Named {
	object := types.NewTypeName(token.NoPos, pkg, name, nil)
	receiver := types.NewNamed(object, types.NewStruct(nil, nil), nil)
	pkg.Scope().Insert(object)
	for getterName, def := range gotextGetter {
		parameterTypes := make([]types.Type, def.MaxArgIndex()+2)
		for idx := range parameterTypes {
			parameterTypes[idx] = anyType
		}
		addTypedMethodWithVariadic(pkg, receiver, getterName, true, parameterTypes...)
	}
	return receiver
}
func addTypedInterface(pkg *types.Package, name string, anyType types.Type) *types.Named {
	methods := make([]*types.Func, 0, len(gotextGetter))
	for getterName, def := range gotextGetter {
		parameterTypes := make([]types.Type, def.MaxArgIndex()+2)
		for idx := range parameterTypes {
			parameterTypes[idx] = anyType
		}
		methods = append(methods, types.NewFunc(
			token.NoPos,
			pkg,
			getterName,
			typedSignature(pkg, nil, true, parameterTypes...),
		))
	}
	interfaceType := types.NewInterfaceType(methods, nil)
	interfaceType.Complete()
	object := types.NewTypeName(token.NoPos, pkg, name, nil)
	named := types.NewNamed(object, interfaceType, nil)
	pkg.Scope().Insert(object)
	return named
}

func buildTypedGoFile(source string) (*GoFile, *ast.File, error) {
	fileSet := token.NewFileSet()
	file, err := goparser.ParseFile(fileSet, "typed.go", source, 0)
	if err != nil {
		return nil, nil, err
	}

	gotextTypes := types.NewPackage("github.com/leonelquinteros/gotext", "gotext")
	anyType := types.NewInterfaceType(nil, nil)
	anyType.Complete()
	addTypedGetterWithVariadic(gotextTypes, "Get", true, types.Typ[types.String], anyType)
	addTypedGetterWithVariadic(gotextTypes, "GetN", true,
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
		anyType,
	)
	addTypedGetterWithVariadic(gotextTypes, "GetD", true,
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
	)
	addTypedGetterWithVariadic(gotextTypes, "GetND", true,
		types.Typ[types.String],
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
		anyType,
	)
	addTypedGetterWithVariadic(gotextTypes, "GetC", true,
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
	)
	addTypedGetterWithVariadic(gotextTypes, "GetNC", true,
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
		types.Typ[types.String],
		anyType,
	)
	addTypedGetterWithVariadic(gotextTypes, "GetDC", true,
		types.Typ[types.String],
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
	)
	addTypedGetterWithVariadic(gotextTypes, "GetNDC", true,
		types.Typ[types.String],
		types.Typ[types.String],
		types.Typ[types.String],
		anyType,
		types.Typ[types.String],
		anyType,
	)
	addTypedReceiver(gotextTypes, "Locale", anyType)
	addTypedReceiver(gotextTypes, "Po", anyType)
	addTypedReceiver(gotextTypes, "Domain", anyType)
	addTypedInterface(gotextTypes, "Translator", anyType)
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
		return nil, nil, typeErr
	}
	importedPackages := map[string]*packages.Package{
		"gotext": {
			PkgPath: "github.com/leonelquinteros/gotext",
		},
		"typed": {
			PkgPath:   "typed",
			Syntax:    []*ast.File{file},
			TypesInfo: typesInfo,
		},
	}
	for _, importSpec := range file.Imports {
		if importSpec.Name != nil {
			importedPackages[importSpec.Name.Name] = importedPackages["gotext"]
		}
	}

	g := &GoFile{
		FilePath:         "typed.go",
		BasePath:         ".",
		Data:             &DomainMap{},
		FileSet:          fileSet,
		ImportedPackages: importedPackages,
	}
	return g, file, nil
}

func newTypedGoFile(t *testing.T, source string) (*GoFile, *ast.File) {
	t.Helper()
	g, file, err := buildTypedGoFile(source)
	if err != nil {
		t.Fatal(err)
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
func TestGoFile_InspectCallExpr_GetterMappings(t *testing.T) {
	const source = `package typed

import gt "github.com/leonelquinteros/gotext"

func example() {
	gt.Get("get-id", "format %s")
	gt.GetN("getn-id", "getn-plural", 2, "format %s")
	gt.GetD("getd-domain", "getd-id", "format %s")
	gt.GetND("getnd-domain", "getnd-id", "getnd-plural", 2, "format %s")
	gt.GetC("getc-id", "getc-context", "format %s")
	gt.GetNC("getnc-id", "getnc-plural", 2, "getnc-context", "format %s")
	gt.GetDC("getdc-domain", "getdc-id", "getdc-context", "format %s")
	gt.GetNDC("getndc-domain", "getndc-id", "getndc-plural", 2, "getndc-context", "format %s")
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	type want struct {
		domain  string
		id      string
		plural  string
		context string
	}
	expected := []want{
		{id: "get-id"},
		{id: "getn-id", plural: "getn-plural"},
		{domain: "getd-domain", id: "getd-id"},
		{domain: "getnd-domain", id: "getnd-id", plural: "getnd-plural"},
		{id: "getc-id", context: `"getc-context"`},
		{id: "getnc-id", plural: "getnc-plural", context: `"getnc-context"`},
		{domain: "getdc-domain", id: "getdc-id", context: `"getdc-context"`},
		{domain: "getndc-domain", id: "getndc-id", plural: "getndc-plural", context: `"getndc-context"`},
	}

	if len(g.Data.Domains) != 5 {
		t.Fatalf("got %d domains, want 5", len(g.Data.Domains))
	}
	for _, expected := range expected {
		domain := g.Data.Domains[expected.domain]
		if domain == nil {
			domain = g.Data.Domains["default"]
		}
		if expected.context == "" {
			translation := domain.Translations[expected.id]
			if translation == nil {
				t.Fatalf("missing translation %q in domain %q", expected.id, expected.domain)
			}
			if translation.MsgIDPlural != expected.plural || translation.Context != "" {
				t.Errorf("translation %q = plural %q/context %q, want plural %q/context empty",
					expected.id, translation.MsgIDPlural, translation.Context, expected.plural)
			}
			continue
		}
		translation := domain.ContextTranslations[expected.context][expected.id]
		if translation == nil {
			t.Fatalf("missing contextual translation %q/%q", expected.context, expected.id)
		}
		if translation.MsgIDPlural != expected.plural || translation.Context != expected.context {
			t.Errorf("translation %q = plural %q/context %q, want plural %q/context %q",
				expected.id, translation.MsgIDPlural, translation.Context, expected.plural, expected.context)
		}
	}
}

func TestGoFile_InspectCallExpr_DotImportsReceiversAndShadowedMethods(t *testing.T) {
	const source = `package typed

import . "github.com/leonelquinteros/gotext"

type local struct{}

func (local) Get(id string, vars ...any) string {
	return id
}

func example() {
	var locale Locale
	var pointer *Locale
	var translator Translator
	Get("dot-id")
	locale.Get("named-id")
	pointer.Get("pointer-id")
	translator.Get("interface-id")
	var localValue local
	localValue.Get("local-id")
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("dot import and receiver calls did not create default domain")
	}
	for _, id := range []string{"dot-id", "named-id", "pointer-id", "interface-id"} {
		if domain.Translations[id] == nil {
			t.Errorf("missing supported translation %q", id)
		}
	}
	if _, ok := domain.Translations["local-id"]; ok {
		t.Error("same-named local method was extracted")
	}
	if len(domain.Translations) != 4 {
		t.Fatalf("got %d translations, want 4", len(domain.Translations))
	}
}

func TestGoFile_InspectCallExpr_ConstantFormsAndRangeMutation(t *testing.T) {
	const source = `package typed

import gt "github.com/leonelquinteros/gotext"

const (
	firstConst = "first"
	inheritedConst
	aliasConst = firstConst + "-alias"
	concatenatedConst = "left" + "right"
	convertedConst = string('x')
)

func dynamicString() string {
	return "dynamic"
}

func example() {
	gt.Get(firstConst)
	gt.Get(inheritedConst)
	gt.Get(aliasConst)
	gt.Get(concatenatedConst)
	gt.Get(convertedConst)

	staticVariable := "static"
	gt.Get(staticVariable)
	afterCall := "before"
	gt.Get(afterCall)
	afterCall = "after"

	dynamicVariable := dynamicString()
	gt.Get(dynamicVariable)
	mutatedVariable := "before mutation"
	mutatedVariable = "after mutation"
	gt.Get(mutatedVariable)

	rangedVariable := "initial"
	for _, rangedVariable = range []string{"range"} {
		gt.Get(rangedVariable)
	}
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("constant calls did not create default domain")
	}
	for _, id := range []string{
		"first",
		"first-alias",
		"leftright",
		"x",
		"static",
		"before",
	} {
		if domain.Translations[id] == nil {
			t.Errorf("missing static translation %q", id)
		}
	}
	for _, id := range []string{"dynamic", "before mutation", "after mutation", "range"} {
		if _, ok := domain.Translations[id]; ok {
			t.Errorf("dynamic or mutated translation %q was extracted", id)
		}
	}
}

func FuzzGoFileInspectCallExprConstantForms(f *testing.F) {
	for _, seed := range []string{
		`"seed"`,
		"`raw seed`",
		`"left" + "right"`,
		`string('x')`,
		`not valid`,
		`"unterminated`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, expression string) {
		if len(expression) > 64<<10 {
			return
		}
		source := `package typed

import gt "github.com/leonelquinteros/gotext"

func example() {
	gt.Get(` + expression + `)
}`
		g, file, err := buildTypedGoFile(source)
		if err != nil {
			return
		}

		var calls []*ast.CallExpr
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if ok && selector.Sel != nil && selector.Sel.Name == "Get" {
				calls = append(calls, call)
			}
			return true
		})
		if len(calls) != 1 || len(calls[0].Args) == 0 {
			return
		}

		call := calls[0]
		g.InspectCallExpr(call)
		literal := g.resolveStringLiteral(call.Args[0], call.Pos(), nil)
		if literal == nil {
			if len(g.Data.Domains) != 0 {
				t.Fatal("dynamic expression produced a translation")
			}
			return
		}
		value, err := strconv.Unquote(literal.Value)
		if err != nil {
			t.Fatalf("resolved literal %q is not a string: %v", literal.Value, err)
		}
		domain := g.Data.Domains["default"]
		if domain == nil {
			t.Fatal("resolved literal did not initialize the default domain")
		}
		translation := domain.Translations[value]
		if translation == nil {
			t.Fatalf("resolved literal %q was not extracted", value)
		}
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

func TestGoFile_ResolveStringLiteral_UsesOwningPackageSyntax(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

var crossFile = "cross-file"

func example() {
	gotext.Get(crossFile)
}`

	g, file := newTypedGoFile(t, source)
	var call *ast.CallExpr
	var getter *ast.Ident
	var declaration *ast.GenDecl
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		declaration = genDecl
		break
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.CallExpr:
			call = node
		case *ast.SelectorExpr:
			if node.Sel != nil && node.Sel.Name == "Get" {
				getter = node.Sel
			}
		}
		return true
	})
	if call == nil || getter == nil || declaration == nil {
		t.Fatal("failed to find getter call and package-level declaration")
	}

	gotextObject := g.GetType(getter)
	if gotextObject == nil || gotextObject.Pkg() == nil {
		t.Fatal("getter has no owning package")
	}
	declIndex := slices.Index(file.Decls, ast.Decl(declaration))
	file.Decls = append(file.Decls[:declIndex], file.Decls[declIndex+1:]...)
	declarationFile, err := goparser.ParseFile(g.FileSet, "declaration.go", `package typed

var crossFile = "cross-file"
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	typesInfo := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	if _, err := (&types.Config{
		Importer: typedASTImporter{pkg: gotextObject.Pkg()},
	}).Check("typed", g.FileSet, []*ast.File{file, declarationFile}, typesInfo); err != nil {
		t.Fatal(err)
	}
	ownerPackage := g.ImportedPackages["typed"]
	if ownerPackage == nil {
		t.Fatal("typed package metadata is missing")
	}
	ownerPackage.Syntax = []*ast.File{file, declarationFile}
	ownerPackage.TypesInfo = typesInfo

	var use, definition *ast.Ident
	if len(call.Args) > 0 {
		use, _ = call.Args[0].(*ast.Ident)
	}
	ast.Inspect(declarationFile, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok && ident.Name == "crossFile" {
			definition = ident
			return false
		}
		return true
	})
	if use == nil || definition == nil {
		t.Fatal("failed to find type-checked getter argument and declaration")
	}
	object := g.GetType(definition)
	if object == nil || object.Pkg() == nil {
		t.Fatal("cross-file identifier has no owning package object")
	}

	unrelatedFile, err := goparser.ParseFile(token.NewFileSet(), "unrelated.go", `package unrelated

var crossFile = "unrelated declaration"

func mutate() {
	crossFile = "unrelated mutation"
}
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	var unrelatedDeclaration, unrelatedMutation *ast.Ident
	ast.Inspect(unrelatedFile, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.ValueSpec:
			if len(node.Names) > 0 {
				unrelatedDeclaration = node.Names[0]
			}
		case *ast.AssignStmt:
			if len(node.Lhs) > 0 {
				unrelatedMutation, _ = node.Lhs[0].(*ast.Ident)
				if unrelatedMutation != nil {
					unrelatedMutation.NamePos = token.Pos(1)
				}
			}
		}
		return true
	})
	if unrelatedDeclaration == nil || unrelatedMutation == nil {
		t.Fatal("failed to find unrelated declarations")
	}
	unrelatedPackage := &packages.Package{
		PkgPath: "example.com/unrelated",
		Syntax:  []*ast.File{unrelatedFile},
		TypesInfo: &types.Info{
			Defs: map[*ast.Ident]types.Object{
				unrelatedDeclaration: object,
			},
			Uses: map[*ast.Ident]types.Object{
				unrelatedMutation: object,
			},
		},
	}

	savedPackages := g.ImportedPackages
	g.ImportedPackages = map[string]*packages.Package{"unrelated": unrelatedPackage}
	if _, _, ok := g.declarationForObject(object); ok {
		t.Fatal("unrelated package declaration was scanned")
	}
	g.ImportedPackages = savedPackages
	g.ImportedPackages["unrelated"] = unrelatedPackage

	if g.isMutatedBefore(object, call.Pos()) {
		t.Fatal("unrelated package mutation was scanned")
	}
	literal := g.resolveStringLiteral(use, call.Pos(), nil)
	if literal == nil || literal.Value != `"cross-file"` {
		t.Fatalf("resolved literal = %v, want %q", literal, `"cross-file"`)
	}

	g.InspectCallExpr(call)
	domain := g.Data.Domains["default"]
	if domain == nil || domain.Translations["cross-file"] == nil {
		t.Fatal("cross-file declaration was not extracted")
	}
}
