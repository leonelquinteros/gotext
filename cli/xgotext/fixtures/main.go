package main

import (
	"errors"
	"fmt"

	"github.com/leonelquinteros/gotext"
	alias "github.com/leonelquinteros/gotext"
	"github.com/leonelquinteros/gotext/cli/xgotext/fixtures/pkg"
)

// Fake object with methods similar to gotext
type Fake struct {
}

// Get by id
func (f Fake) Get(id int) int {
	return 42
}

// Fake object with same methods as gotext
type Fake2 struct {
}

// Get by str
func (f Fake2) Get(s string) string {
	return s
}

// Type alias
type MyPo = *gotext.Po

// Function return type
var NL = func() *gotext.Locale {
	return gotext.NewLocale("/path/to/locales/root/dir", "es_UY")
}

func main() {
	// Configure package
	gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

	// Translate text from default domain
	fmt.Println(gotext.Get("My text on 'domain-name' domain"))
	// same as before
	fmt.Println(gotext.Get("My text on 'domain-name' domain"))

	// unsupported function call
	trStr := "some string to translate"
	fmt.Println(gotext.Get(trStr))

	// same with alias package name
	fmt.Println(alias.Get("alias call"))

	// Special strings
	fmt.Println(gotext.Get(`string with backquotes`))
	fmt.Println(gotext.Get("string ending with EOL\n"))
	fmt.Println(gotext.Get("string with\nmultiple\nEOL"))
	fmt.Println(gotext.Get(`raw string with\nmultiple\nEOL`))
	fmt.Println(gotext.Get(`multi
line
string`))
	fmt.Println(gotext.Get(`multi
line
string
ending with
EOL`))
	fmt.Println(gotext.Get("multline\nending with EOL\n"))

	// Translate text from a different domain without reconfigure
	fmt.Println(gotext.GetD("domain2", "Another text on a different domain"))

	// Create Locale with library path and language code
	l := NL()

	// dummy call
	dummy(l)

	// Load domain '/path/to/locales/root/dir/es_UY/default.po'
	l.AddDomain("translations")
	l.SetDomain("translations")

	// Translate text from domain
	fmt.Println(l.GetD("translations", "Translate this"))

	// Get plural translations
	l.GetN("Singular", "Plural", 4)
	num := 17
	l.GetN("SingularVar", "PluralVar", num)

	l.GetDC("domain2", "string", "ctx")
	l.GetNDC("translations", "ndc", "ndcs", 7, "NDC-CTX")

	// try fake structs
	f := Fake{}
	_ = f.Get(3)

	f2 := Fake2{}
	_ = f2.Get("3")

	// use translator of sub object
	t := pkg.Translate{}
	t.L.Get("translate package")
	t.S.L.Get("translate sub package")

	// redefine alias with fake struct
	alias := Fake2{}
	_ = alias.Get("3")

	err := errors.New("test")
	fmt.Print(err.Error())

	// Get from type alias
	var po MyPo
	_ = po.Get("type alias")

	// Locale constructor call
	NL().Get("locale constructor call")
}

// dummy function
func dummy(locale *gotext.Locale) {
	locale.Get("inside dummy")
}
