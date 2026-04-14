# Plural Forms with gotext

`gotext` fully supports GNU Gettext plural forms. This guide explains how to use plural forms in your Go code and PO files.

## 1. Using Plural Forms in Go

To handle plurals, use the `GetN` or `GetND` functions (the `N` stands for plural-aware).

### Package-Level API

```go
package main

import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    gotext.Configure("locales", "en_US", "default")

    apples := 5
    // GetN(singular msgid, plural msgid, quantity, formatting arguments...)
    fmt.Println(gotext.GetN("I have one apple.", "I have %d apples.", apples, apples))
}
```

## 2. Setting Up Your PO Files

Plural forms rely on the `Plural-Forms` header in your PO file. This header defines the number of plural forms for the language and the rule (a C-like expression) to select the correct form based on the quantity `n`.

### Example: English (2 forms: singular/plural)
```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "I have one apple."
msgid_plural "I have %d apples."
msgstr[0] "I have one apple."
msgstr[1] "I have %d apples."
```

### Example: Polish (3 forms)
```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=3; plural=(n==1 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2);\n"

msgid "I have one apple."
msgid_plural "I have %d apples."
msgstr[0] "Mam jedno jabłko."
msgstr[1] "Mam %d jabłka."
msgstr[2] "Mam %d jabłek."
```

## 3. How it Works

1.  **Header Parsing**: When `gotext` loads a PO file, it parses the `Plural-Forms` header.
2.  **Expression Evaluation**: When `GetN` is called, `gotext` evaluates the plural expression with the provided `n`.
3.  **Result Indexing**: The evaluation result (0, 1, 2, etc.) is used as an index to select the correct `msgstr[n]` from the translation entry.

For more information on plural form rules for different languages, see the [GNU Gettext manual](https://www.gnu.org/savannah-checkouts/gnu/gettext/manual/html_node/Plural-forms.html).
