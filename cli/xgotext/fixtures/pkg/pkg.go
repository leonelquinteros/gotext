package pkg

import "github.com/donseba/gotext"

type SubTranslate struct {
	L gotext.Locale
}

type Translate struct {
	L gotext.Locale
	S SubTranslate
}

func test() {
	gotext.Get("inside sub package")
}
