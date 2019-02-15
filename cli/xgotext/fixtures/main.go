package main

import (
	"fmt"

	"github.com/leonelquinteros/gotext"
)

// Fake object with methods similar to gotext
type Fake struct {
}

// Get by id
func (f Fake) Get(id int) int {
	return 42
}

func main() {
	// Configure package
	gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

	// Translate text from default domain
	fmt.Println(gotext.Get("My text on 'domain-name' domain"))

	// Translate text from a different domain without reconfigure
	fmt.Println(gotext.GetD("domain2", "Another text on a different domain"))

	// Create Locale with library path and language code
	l := gotext.NewLocale("/path/to/locales/root/dir", "es_UY")

	// Load domain '/path/to/locales/root/dir/es_UY/default.po'
	l.AddDomain("default")

	// Translate text from domain
	fmt.Println(l.GetD("translations", "Translate this"))

	// Get plural translations
	l.GetN("Singular", "Plural", 4)
	num := 17
	l.GetN("SingularVar", "PluralVar", num)

	l.GetDC("domain", "string", "ctx")
	l.GetNDC("translations", "ndc", "ndcs", 7, "NDC-CTX")

	f := Fake{}
	f.Get(3)
}
