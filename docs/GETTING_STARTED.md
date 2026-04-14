# Getting Started with gotext

Welcome to `gotext`! This guide will help you get up and running with GNU Gettext in Go in just a few minutes.

## 1. Installation

Add `gotext` to your project using `go get`:

```bash
go get github.com/leonelquinteros/gotext
```

## 2. Basic Example: Package-Level API

For simple applications, you can use the package-level functions directly. This is great for apps with a single primary language and domain.

### Project Structure
Assume you have a `locales` directory like this:
```
/locales/en_US/default.po
```

### main.go
```go
package main

import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // 1. Configure the package
    // Path to locales, language code, and default domain
    gotext.Configure("locales", "en_US", "default")

    // 2. Use it!
    fmt.Println(gotext.Get("Hello, world!"))
    
    // 3. Use it with variables
    name := "Gopher"
    fmt.Println(gotext.Get("Hello, %s!", name))
}
```

## 3. Multiple Languages: Using the Locale Object

For more complex applications, use the `Locale` object to manage multiple languages independently.

```go
package main

import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // Create a new Locale for Spanish (Uruguay)
    l := gotext.NewLocale("locales", "es_UY")

    // Load the 'default' domain
    l.AddDomain("default")

    // Translate!
    fmt.Println(l.Get("Translate this"))
}
```

## 4. Next Steps
- Learn how to extract strings from your code using [xgotext](xgotext.md).
- Explore [Plural Forms](PLURALS.md) (coming soon).
- See more examples in the [README](../README.md#usage-examples).
