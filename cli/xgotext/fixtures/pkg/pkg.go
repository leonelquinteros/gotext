package pkg

import "github.com/leonelquinteros/gotext"

// SubTranslate is a sub package for testing
type SubTranslate struct {
	L gotext.Locale
}

// Translate is a struct for testing
type Translate struct {
	L gotext.Locale
	S SubTranslate
}

func test() {
	gotext.Get("inside sub package")
}
