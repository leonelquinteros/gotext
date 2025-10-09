package main

import (
	"fmt"

	"github.com/leonelquinteros/gotext"
)

func main() {
	// with format arguments
	fmt.Println(gotext.Get("simple %s", "param"))
	fmt.Println(gotext.GetD("domain", "simple %s", "param"))
	fmt.Println(gotext.GetC("simple %s", "ctx", "param"))
	fmt.Println(gotext.GetDC("domain", "simple %s", "ctx", "param"))
	fmt.Println(gotext.GetN("singular %s", "plural %s", 2, "param"))
	fmt.Println(gotext.GetND("domain", "singular %s", "plural %s", 2, "param"))
	fmt.Println(gotext.GetNC("singular %s", "plural %s", 2, "ctx", "param"))
	fmt.Println(gotext.GetNDC("domain", "singular %s", "plural %s", 2, "ctx", "param"))

	// without format arguments
	fmt.Println(gotext.Get("simple"))
	fmt.Println(gotext.GetD("domain", "simple"))
	fmt.Println(gotext.GetC("simple", "ctx"))
	fmt.Println(gotext.GetDC("domain", "simple", "ctx"))
	fmt.Println(gotext.GetN("singular", "plural", 2))
	fmt.Println(gotext.GetND("domain", "singular", "plural", 2))
	fmt.Println(gotext.GetNC("singular", "plural", 2, "ctx"))
	fmt.Println(gotext.GetNDC("domain", "singular", "plural", 2, "ctx"))

}
