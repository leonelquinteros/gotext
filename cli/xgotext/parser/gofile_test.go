package parser

import (
	"fmt"
	"go/ast"
	"go/constant"
	goparser "go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

const typedGotextPackagePath = "github.com/leonelquinteros/gotext"

type typedASTImporter struct {
	packages map[string]*types.Package
}

func (i typedASTImporter) Import(path string) (*types.Package, error) {
	pkg := i.packages[path]
	if pkg == nil {
		return nil, fmt.Errorf("typed fixture importer has no package %q", path)
	}
	return pkg, nil
}

type typedGetterSpec struct {
	name           string
	parameterTypes []types.Type
}

func typedGetterSpecs(anyType types.Type) []typedGetterSpec {
	stringType := types.Typ[types.String]
	intType := types.Typ[types.Int]
	return []typedGetterSpec{
		{name: "Get", parameterTypes: []types.Type{stringType, anyType}},
		{name: "GetN", parameterTypes: []types.Type{stringType, stringType, intType, anyType}},
		{name: "GetD", parameterTypes: []types.Type{stringType, stringType, anyType}},
		{name: "GetND", parameterTypes: []types.Type{stringType, stringType, stringType, intType, anyType}},
		{name: "GetC", parameterTypes: []types.Type{stringType, stringType, anyType}},
		{name: "GetNC", parameterTypes: []types.Type{stringType, stringType, intType, stringType, anyType}},
		{name: "GetDC", parameterTypes: []types.Type{stringType, stringType, stringType, anyType}},
		{name: "GetNDC", parameterTypes: []types.Type{stringType, stringType, stringType, intType, stringType, anyType}},
	}
}

func typedReceiverGetterSpecs(anyType types.Type, name string) []typedGetterSpec {
	specs := typedGetterSpecs(anyType)
	switch name {
	case "Locale", "Po":
		return specs
	case "Domain", "Translator":
		return []typedGetterSpec{specs[0], specs[1], specs[4], specs[5]}
	default:
		panic("unknown typed receiver " + name)
	}
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

func addTypedGetterWithVariadic(pkg *types.Package, spec typedGetterSpec) {
	pkg.Scope().Insert(types.NewFunc(
		token.NoPos,
		pkg,
		spec.name,
		typedSignature(pkg, nil, true, spec.parameterTypes...),
	))
}

func addTypedReceiver(pkg *types.Package, name string, specs []typedGetterSpec) *types.Named {
	object := types.NewTypeName(token.NoPos, pkg, name, nil)
	receiver := types.NewNamed(object, types.NewStruct(nil, nil), nil)
	pkg.Scope().Insert(object)
	for _, spec := range specs {
		receiver.AddMethod(types.NewFunc(
			token.NoPos,
			pkg,
			spec.name,
			typedSignature(pkg, receiver, true, spec.parameterTypes...),
		))
	}
	return receiver
}

func addTypedHeaderMap(pkg *types.Package) {
	stringType := types.Typ[types.String]
	object := types.NewTypeName(token.NoPos, pkg, "HeaderMap", nil)
	headerMap := types.NewNamed(
		object,
		types.NewMap(stringType, types.NewSlice(stringType)),
		nil,
	)
	pkg.Scope().Insert(object)
	headerMap.AddMethod(types.NewFunc(
		token.NoPos,
		pkg,
		"Get",
		typedSignature(pkg, headerMap, false, stringType),
	))
}

func addTypedInterface(pkg *types.Package, name string, specs []typedGetterSpec) *types.Named {
	methods := make([]*types.Func, 0, len(specs))
	for _, spec := range specs {
		methods = append(methods, types.NewFunc(
			token.NoPos,
			pkg,
			spec.name,
			typedSignature(pkg, nil, true, spec.parameterTypes...),
		))
	}
	interfaceType := types.NewInterfaceType(methods, nil)
	interfaceType.Complete()
	object := types.NewTypeName(token.NoPos, pkg, name, nil)
	named := types.NewNamed(object, interfaceType, nil)
	pkg.Scope().Insert(object)
	return named
}

func typedTypesInfo() *types.Info {
	return &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
}

func buildTypedGoFile(source string) (*GoFile, *ast.File, error) {
	fileSet := token.NewFileSet()
	file, err := goparser.ParseFile(fileSet, "typed.go", source, 0)
	if err != nil {
		return nil, nil, err
	}

	gotextTypes := types.NewPackage(typedGotextPackagePath, "gotext")
	anyType := types.NewInterfaceType(nil, nil)
	anyType.Complete()
	specs := typedGetterSpecs(anyType)
	for _, spec := range specs {
		addTypedGetterWithVariadic(gotextTypes, spec)
	}
	addTypedReceiver(gotextTypes, "Locale", typedReceiverGetterSpecs(anyType, "Locale"))
	addTypedReceiver(gotextTypes, "Po", typedReceiverGetterSpecs(anyType, "Po"))
	addTypedReceiver(gotextTypes, "Domain", typedReceiverGetterSpecs(anyType, "Domain"))
	addTypedInterface(gotextTypes, "Translator", typedReceiverGetterSpecs(anyType, "Translator"))
	addTypedHeaderMap(gotextTypes)
	gotextTypes.MarkComplete()

	typesInfo := typedTypesInfo()
	ownerTypes, typeErr := (&types.Config{
		Importer: typedASTImporter{packages: map[string]*types.Package{
			typedGotextPackagePath: gotextTypes,
		}},
	}).Check("typed", fileSet, []*ast.File{file}, typesInfo)
	if typeErr != nil {
		return nil, nil, typeErr
	}

	gotextPackage := &packages.Package{
		Name:    "gotext",
		PkgPath: typedGotextPackagePath,
		Types:   gotextTypes,
	}
	ownerPackage := &packages.Package{
		Name:      "typed",
		PkgPath:   "typed",
		Syntax:    []*ast.File{file},
		Types:     ownerTypes,
		TypesInfo: typesInfo,
	}
	importedPackages := map[string]*packages.Package{
		"gotext": gotextPackage,
		"typed":  ownerPackage,
	}
	for _, importSpec := range file.Imports {
		if importSpec.Name != nil {
			importedPackages[importSpec.Name.Name] = gotextPackage
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

	type expectedTranslation struct {
		domain, id, plural, context, location string
		contextual                            bool
	}
	expected := []expectedTranslation{
		{id: "get-id", location: "typed.go:6"},
		{id: "getn-id", plural: "getn-plural", location: "typed.go:7"},
		{domain: "getd-domain", id: "getd-id", location: "typed.go:8"},
		{domain: "getnd-domain", id: "getnd-id", plural: "getnd-plural", location: "typed.go:9"},
		{id: "getc-id", context: "getc-context", contextual: true, location: "typed.go:10"},
		{id: "getnc-id", plural: "getnc-plural", context: "getnc-context", contextual: true, location: "typed.go:11"},
		{domain: "getdc-domain", id: "getdc-id", context: "getdc-context", contextual: true, location: "typed.go:12"},
		{domain: "getndc-domain", id: "getndc-id", plural: "getndc-plural", context: "getndc-context", contextual: true, location: "typed.go:13"},
	}

	if len(g.Data.Domains) != 5 {
		t.Fatalf("got %d domains, want 5", len(g.Data.Domains))
	}
	for _, want := range expected {
		domain := g.Data.Domains[want.domain]
		if domain == nil {
			domain = g.Data.Domains["default"]
		}
		var translation *Translation
		if want.contextual {
			translation = domain.ContextTranslations[want.context][want.id]
		} else {
			translation = domain.Translations[want.id]
		}
		if translation == nil {
			t.Fatalf("missing translation %q in domain %q", want.id, want.domain)
		}
		if translation.MsgID != want.id ||
			translation.MsgIDPlural != want.plural ||
			translation.Context != want.context ||
			translation.HasContext != want.contextual ||
			len(translation.SourceLocations) != 1 ||
			translation.SourceLocations[0] != want.location {
			t.Errorf("translation = %#v, want id=%q plural=%q context=%q contextual=%t location=%q",
				translation, want.id, want.plural, want.context, want.contextual, want.location)
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
	var headers HeaderMap
	Get("dot-id")
	locale.Get("named-id")
	pointer.Get("pointer-id")
	translator.Get("interface-id")
	headers.Get("Content-Type")
	var localValue local
	localValue.Get("local-id")
}
`

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
	if _, ok := domain.Translations["Content-Type"]; ok {
		t.Error("HeaderMap.Get key was extracted as a translation")
	}
	if len(domain.Translations) != 4 {
		t.Fatalf("got %d translations, want 4", len(domain.Translations))
	}
}

func TestGoFile_InspectCallExpr_MethodExpressions(t *testing.T) {
	const source = `package typed

import gt "github.com/leonelquinteros/gotext"

func example() {
	var locale gt.Locale
	(*gt.Locale).Get(&locale, "method-expression")
	(*gt.Locale).GetC(&locale, "method-context-id", "method-context")
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("method expressions did not create the default domain")
	}
	if translation := domain.Translations["method-expression"]; translation == nil ||
		translation.MsgID != "method-expression" || translation.HasContext {
		t.Fatalf("method-expression translation = %#v, want decoded uncontextualized metadata", translation)
	}
	translation := domain.ContextTranslations["method-context"]["method-context-id"]
	if translation == nil ||
		translation.MsgID != "method-context-id" ||
		translation.Context != "method-context" ||
		!translation.HasContext {
		t.Fatalf("method-context translation = %#v, want decoded contextual metadata", translation)
	}
}

func TestGoFile_InspectCallExpr_ContextDecodingAndEmptyContext(t *testing.T) {
	const source = `package typed

import gt "github.com/leonelquinteros/gotext"

func example() {
	gt.GetC("same-context", ` + "`ctx`" + `, "format")
	gt.GetC("same-context", "ctx", "format")
	gt.GetC("same-context", "\u0063tx", "format")
	gt.GetC("empty-context", "", "format")
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("context calls did not create the default domain")
	}
	if len(domain.Translations) != 0 {
		t.Fatalf("got %d uncontextualized translations, want none", len(domain.Translations))
	}
	if len(domain.ContextTranslations) != 2 {
		t.Fatalf("got %d context buckets, want decoded ctx and empty context", len(domain.ContextTranslations))
	}
	contextual := domain.ContextTranslations["ctx"]["same-context"]
	if contextual == nil ||
		contextual.Context != "ctx" ||
		!contextual.HasContext ||
		len(contextual.SourceLocations) != 3 {
		t.Fatalf("decoded context translation = %#v, want one entry with three source locations", contextual)
	}
	emptyContext := domain.ContextTranslations[""]["empty-context"]
	if emptyContext == nil ||
		emptyContext.Context != "" ||
		!emptyContext.HasContext ||
		len(emptyContext.SourceLocations) != 1 {
		t.Fatalf("explicit empty context translation = %#v, want contextual empty metadata", emptyContext)
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
	expected := map[string]string{
		"first":       "",
		"first-alias": "",
		"leftright":   "",
		"x":           "",
		"static":      "",
		"before":      "",
	}
	if len(domain.Translations) != len(expected) {
		t.Fatalf("got %d translations, want %d", len(domain.Translations), len(expected))
	}
	for id, plural := range expected {
		translation := domain.Translations[id]
		if translation == nil {
			t.Fatalf("missing static translation %q", id)
		}
		if translation.MsgID != id || translation.MsgIDPlural != plural ||
			translation.Context != "" || translation.HasContext {
			t.Errorf("translation %q = %#v, want id=%q plural=%q without context",
				id, translation, id, plural)
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

		var call *ast.CallExpr
		ast.Inspect(file, func(node ast.Node) bool {
			candidate, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := candidate.Fun.(*ast.SelectorExpr)
			if ok && selector.Sel != nil && selector.Sel.Name == "Get" {
				call = candidate
			}
			return true
		})
		if call == nil || len(call.Args) == 0 {
			return
		}

		typeInfo := g.ImportedPackages["typed"].TypesInfo.Types[call.Args[0]]
		expected, isConstant := "", false
		if typeInfo.Value != nil && typeInfo.Value.Kind() == constant.String {
			expected = constant.StringVal(typeInfo.Value)
			isConstant = true
		}

		g.InspectCallExpr(call)
		if !isConstant {
			if len(g.Data.Domains) != 0 {
				t.Fatal("dynamic expression produced a translation")
			}
			return
		}

		domain := g.Data.Domains["default"]
		if domain == nil {
			t.Fatal("constant expression did not initialize the default domain")
		}
		if len(domain.Translations) != 1 {
			t.Fatalf("got %d translations, want one for %q", len(domain.Translations), expected)
		}
		translation := domain.Translations[expected]
		if translation == nil || translation.MsgID != expected ||
			translation.Context != "" || translation.HasContext || translation.MsgIDPlural != "" {
			t.Fatalf("translation = %#v, want literal %q without plural or context", translation, expected)
		}
	})
}

func makeTypedASTCycle(t *testing.T, g *GoFile, file *ast.File) {
	t.Helper()
	owner := g.ImportedPackages["typed"]
	if owner == nil || owner.TypesInfo == nil {
		t.Fatal("typed owner package has no type information")
	}

	declarations := make(map[string]*ast.ValueSpec, 2)
	objects := make(map[string]types.Object, 2)
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for _, name := range spec.Names {
			if name.Name != "cyclicA" && name.Name != "cyclicB" {
				continue
			}
			object := owner.TypesInfo.Defs[name]
			if object == nil || object.Pkg() == nil || object.Pkg().Path() != "typed" {
				t.Fatalf("%s is not a type-checked owner-package object", name.Name)
			}
			declarations[name.Name] = spec
			objects[name.Name] = object
		}
		return true
	})

	references := make(map[string]*ast.Ident, 2)
	ast.Inspect(file, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		for name, object := range objects {
			if ident.Name == name && owner.TypesInfo.Uses[ident] == object {
				references[name] = ident
			}
		}
		return true
	})

	for _, name := range []string{"cyclicA", "cyclicB"} {
		if declarations[name] == nil || references[name] == nil {
			t.Fatalf("failed to find type-checked cycle node %s", name)
		}
	}
	declarations["cyclicA"].Values = []ast.Expr{references["cyclicB"]}
	declarations["cyclicB"].Values = []ast.Expr{references["cyclicA"]}
}

func TestGoFile_InspectCallExpr_TypedASTMutation(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

func example() {
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

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("mutation test did not initialize the default domain")
	}
	translations := domain.Translations
	if len(translations) != 2 {
		t.Fatalf("got %d translations, want 2", len(translations))
	}
	before := translations["message before mutation"]
	if before == nil || before.MsgID != "message before mutation" {
		t.Errorf("unmodified identifier before its call = %#v, want literal metadata", before)
	}
	afterCall := translations["message before later mutation"]
	if afterCall == nil || afterCall.MsgID != "message before later mutation" {
		t.Errorf("identifier mutated after its call = %#v, want literal metadata", afterCall)
	}
	if _, ok := translations["message after mutation"]; ok {
		t.Error("identifier mutated before its call was extracted")
	}
}

func TestGoFile_InspectCallExpr_TypedASTPointerEscapeRejectsStaleValue(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

func example() {
	gotext.Get("control")
	pointerValue := "stale-pointer"
	pointer := &pointerValue
	*pointer = "actual-pointer"
	gotext.Get(pointerValue)
}
`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("pointer escape test did not create the default domain")
	}
	if domain.Translations["control"] == nil {
		t.Fatal("control translation was not extracted")
	}
	if _, ok := domain.Translations["stale-pointer"]; ok {
		t.Error("address-escaped variable initializer was extracted")
	}
	if _, ok := domain.Translations["actual-pointer"]; ok {
		t.Error("address-escaped variable mutation was extracted")
	}
}

func TestGoFile_InspectCallExpr_LoopCarriedMutation(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

func example() {
	value := "stale-loop"
	for i := 0; i < 2; i++ {
		gotext.Get(value)
		value = "changed-loop"
	}
	stable := "straight-line"
	gotext.Get(stable)
	stable = "later"
}`
	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)
	domain := g.Data.Domains["default"]
	if domain == nil || domain.Translations["straight-line"] == nil {
		t.Fatal("unmutated straight-line argument was not extracted")
	}
	if domain.Translations["stale-loop"] != nil || domain.Translations["changed-loop"] != nil {
		t.Fatal("loop-carried variable was extracted as a fixed message")
	}
}

func TestGoFile_InspectCallExpr_TypedASTGlobalHelperMutationRejectsStaleValue(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

var helperValue = "stale-global"

func example() {
	updateHelperValue()
	gotext.Get("control")
	gotext.Get(helperValue)
}

func updateHelperValue() {
	helperValue = "actual-global"
}`

	g, file := newTypedGoFile(t, source)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("global helper mutation test did not create the default domain")
	}
	if domain.Translations["control"] == nil {
		t.Fatal("control translation was not extracted")
	}
	if _, ok := domain.Translations["stale-global"]; ok {
		t.Error("global variable with helper mutation was extracted")
	}
	if _, ok := domain.Translations["actual-global"]; ok {
		t.Error("global helper assignment was extracted as a message")
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
	gotext.GetNC("id", "plural", 2, "context after cycle", cyclicB, contextAfterCycle)
	gotext.Get(later)
}`

	g, file := newTypedGoFile(t, source)
	makeTypedASTCycle(t, g, file)
	inspectTypedCalls(g, file)

	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("cycle test did not initialize the default domain")
	}
	if len(domain.Translations) != 1 {
		t.Fatalf("got %d uncontexted translations, want one", len(domain.Translations))
	}
	if _, ok := domain.Translations["cyclicA"]; ok {
		t.Error("cyclic identifier was extracted")
	}
	if _, ok := domain.Translations["cyclicB"]; ok {
		t.Error("cyclic identifier was extracted as a later argument")
	}
	later := domain.Translations["later call"]
	if later == nil || later.MsgID != "later call" {
		t.Fatalf("later translation = %#v, want literal call metadata", later)
	}

	contextTranslations := domain.ContextTranslations["context after cycle"]
	if len(contextTranslations) != 1 {
		t.Fatalf("got %d contextual translations, want one", len(contextTranslations))
	}
	translation := contextTranslations["id"]
	if translation == nil {
		t.Fatal("later arguments were not extracted after a cyclic identifier")
	}
	if translation.MsgID != "id" ||
		translation.MsgIDPlural != "plural" ||
		translation.Context != "context after cycle" ||
		!translation.HasContext {
		t.Errorf("contextual translation = %#v, want id/plural/context metadata", translation)
	}
}

func TestGoFile_InspectCallExpr_UsesOwningPackageSyntax(t *testing.T) {
	const source = `package typed

import "github.com/leonelquinteros/gotext"

var crossFile = "temporary declaration"

func example() {
	gotext.Get(crossFile)
}`

	g, file := newTypedGoFile(t, source)
	var call *ast.CallExpr
	var declaration *ast.GenDecl
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok &&
				selector.Sel != nil && selector.Sel.Name == "Get" {
				call = node
			}
		case *ast.GenDecl:
			if node.Tok == token.VAR {
				declaration = node
			}
		}
		return true
	})
	if call == nil || declaration == nil {
		t.Fatal("failed to find typed getter call and temporary declaration")
	}

	declIndex := -1
	for index, decl := range file.Decls {
		if decl == declaration {
			declIndex = index
			break
		}
	}
	if declIndex < 0 {
		t.Fatal("temporary declaration is not part of the owner file")
	}
	file.Decls = append(file.Decls[:declIndex], file.Decls[declIndex+1:]...)

	declarationFile, err := goparser.ParseFile(g.FileSet, "declaration.go", `package typed

var crossFile = "cross-file"
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	gotextPackage := g.ImportedPackages["gotext"]
	if gotextPackage == nil || gotextPackage.Types == nil {
		t.Fatal("typed fixture has no type-checked gettext package")
	}
	ownerInfo := typedTypesInfo()
	ownerTypes, err := (&types.Config{
		Importer: typedASTImporter{packages: map[string]*types.Package{
			typedGotextPackagePath: gotextPackage.Types,
		}},
	}).Check("typed", g.FileSet, []*ast.File{file, declarationFile}, ownerInfo)
	if err != nil {
		t.Fatal(err)
	}
	ownerPackage := g.ImportedPackages["typed"]
	ownerPackage.Syntax = []*ast.File{file, declarationFile}
	ownerPackage.Types = ownerTypes
	ownerPackage.TypesInfo = ownerInfo

	var definition *ast.Ident
	ast.Inspect(declarationFile, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok && ident.Name == "crossFile" && ownerInfo.Defs[ident] != nil {
			definition = ident
			return false
		}
		return true
	})
	if definition == nil {
		t.Fatal("cross-file definition is not type-checked")
	}
	ownerObject := ownerInfo.Defs[definition]
	if ownerObject == nil || ownerObject.Pkg() == nil || ownerObject.Pkg().Path() != "typed" {
		t.Fatal("cross-file definition has no typed owner package")
	}
	var use *ast.Ident
	ast.Inspect(file, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok && ident.Name == "crossFile" && ownerInfo.Uses[ident] == ownerObject {
			use = ident
			return false
		}
		return true
	})
	if use == nil {
		t.Fatal("cross-file use is not linked to the owner-package object")
	}
	// A cross-file identifier cannot retain the parser's same-file ast.Object.
	// Force production lookup through the jointly type-checked owner package.
	use.Obj = nil

	foreignFileSet := token.NewFileSet()
	foreignFile, err := goparser.ParseFile(foreignFileSet, "unrelated.go", `package unrelated

var crossFile = "foreign declaration"

func mutate() {
	crossFile = "foreign mutation"
}
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	foreignInfo := typedTypesInfo()
	foreignTypes, err := (&types.Config{}).Check(
		"example.com/unrelated",
		foreignFileSet,
		[]*ast.File{foreignFile},
		foreignInfo,
	)
	if err != nil {
		t.Fatal(err)
	}
	var foreignDefinition, foreignMutation *ast.Ident
	ast.Inspect(foreignFile, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if !ok || ident.Name != "crossFile" {
			return true
		}
		if foreignInfo.Defs[ident] != nil {
			foreignDefinition = ident
		}
		if foreignInfo.Uses[ident] != nil {
			foreignMutation = ident
		}
		return true
	})
	if foreignDefinition == nil || foreignMutation == nil {
		t.Fatal("foreign fixture objects are not type-checked")
	}
	foreignObject := foreignInfo.Defs[foreignDefinition]
	if foreignObject == nil || foreignObject == ownerObject ||
		foreignObject.Pkg() == nil || foreignObject.Pkg().Path() != "example.com/unrelated" ||
		foreignInfo.Uses[foreignMutation] != foreignObject {
		t.Fatal("foreign fixture does not contain an independent package object")
	}
	g.ImportedPackages["unrelated"] = &packages.Package{
		Name:      "unrelated",
		PkgPath:   "example.com/unrelated",
		Syntax:    []*ast.File{foreignFile},
		Types:     foreignTypes,
		TypesInfo: foreignInfo,
	}

	g.InspectCallExpr(call)
	domain := g.Data.Domains["default"]
	if domain == nil {
		t.Fatal("cross-file call did not initialize the default domain")
	}
	if len(domain.Translations) != 1 {
		t.Fatalf("got %d translations, want one owner-package translation", len(domain.Translations))
	}
	translation := domain.Translations["cross-file"]
	if translation == nil || translation.MsgID != "cross-file" ||
		len(translation.SourceLocations) != 1 || translation.SourceLocations[0] != "typed.go:8" {
		t.Fatalf("translation = %#v, want owner declaration value and call location", translation)
	}
	if _, ok := domain.Translations["foreign declaration"]; ok {
		t.Fatal("foreign-package declaration was extracted")
	}
}
