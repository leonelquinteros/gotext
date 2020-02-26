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

	// Translate text from a different domain without reconfigure
	fmt.Println(gotext.GetD("domain2", "Another text on a different domain"))

	// Create Locale with library path and language code
	l := gotext.NewLocale("/path/to/locales/root/dir", "es_UY")

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
	f.Get(3)

	f2 := Fake2{}
	f2.Get("3")

	// use translator of sub object
	t := pkg.Translate{}
	t.L.Get("translate package")
	t.S.L.Get("translate sub package")

	// redefine alias with fake struct
	alias := Fake2{}
	alias.Get("3")

	err := errors.New("test")
	fmt.Print(err.Error())
}

// dummy function
func dummy(locale *gotext.Locale) {
	locale.Get("inside dummy")
}
