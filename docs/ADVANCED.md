# Advanced Usage

This guide covers the more advanced features of `gotext`: embedding translation files in your binary, named formatting, serialization for caching, language fallback behavior, and custom translation backends.

## 1. Embedding Files with `embed.FS`

Instead of loading `.po`/`.mo` files from disk, you can read them from an in-memory filesystem.

**`NewLocaleFS(lang string, filesystem fs.FS)`** creates a `Locale` whose translation files live at the root of the given filesystem:

```go
package main

import (
    "embed"
    "fmt"

    "github.com/leonelquinteros/gotext"
)

//go:embed locales
var localeFS embed.FS

func main() {
    l := gotext.NewLocaleFS("en_US", localeFS)
    l.AddDomain("default")

    fmt.Println(l.Get("Hello, world!"))
}
```

**`NewLocaleFSWithPath(lang string, filesystem fs.FS, path string)`** works the same but places the locale files under a folder *inside* the filesystem — useful when the language directories are nested below the embed root:

```go
//go:embed fixtures
var fixtureFS embed.FS

func main() {
    // Files live under fixtures/en_US/LC_MESSAGES/default.po
    l := gotext.NewLocaleFSWithPath("en_US", fixtureFS, "fixtures")
    l.AddDomain("default")
    fmt.Println(l.Get("Hello, world!"))
}
```

`NewPoFS` and `NewMoFS` provide the same filesystem support for standalone `Po`/`Mo` objects.

## 2. Named Formatting

The `Get`/`GetN`/`GetC` family inserts variables using standard `fmt` syntax (`%s`, `%d`, ...). If you prefer named placeholders, use the `Sprintf` helper with `%(name)s` style arguments:

```go
fmt.Println(gotext.Sprintf("%(name)s is Type %(type)s", map[string]any{
    "name": "Gotext",
    "type": "struct",
}))
// Output: Gotext is Type struct
```

`NPrintf` is the same helper writing to standard output. Since translation messages go through the exact same formatting pipeline, you can combine both:

```go
msg := gotext.Get("Hello %(name)s!")
fmt.Println(gotext.Sprintf(msg, map[string]any{"name": "Gopher"}))
```

## 3. Serialization for Caching

`Locale` objects implement the `encoding.BinaryMarshaler`/`encoding.BinaryUnmarshaler` interfaces (`MarshalBinary`/`UnmarshalBinary`). This lets you parse translation files once at build time and load the serialized result at runtime — useful in servers or CLIs where reading and parsing many files on every startup is too slow.

```go
// Build/cache step: serialize once
l := gotext.NewLocale("locales", "en_US")
l.AddDomain("default")

data, err := l.MarshalBinary()
if err != nil {
    log.Fatal(err)
}

// Runtime step: restore from bytes
restored := new(gotext.Locale)
if err := restored.UnmarshalBinary(data); err != nil {
    log.Fatal(err)
}

fmt.Println(restored.Get("Hello, world!"))
```

The serialized bytes are self-contained (they include all parsed domains and translations), so they can be stored anywhere — a file, a database, a shared cache.

To use manually built `Locale` objects at the package level — for example, built from in-memory `Po` objects or from an `embed.FS` — replace the package configuration with `SetLocales`:

```go
l := gotext.NewLocaleFS("de_DE", localeFS)
l.AddDomain("default")
gotext.SetLocales([]*gotext.Locale{l})
```

`SetLocales` derives the package-level language, library path, and domain from the first locale. `SetStorage` is a deprecated alias for a single locale.

## 4. Language and Domain Fallback

### Simplified Locales

Locale codes are normalized automatically: `en_US.UTF-8`, `en_US@euro`, and `en_US` are all equivalent. Use `SimplifiedLocale` directly if you need the normalized form.

### Multiple Languages

You can supply a colon-separated list of languages (package level) or configure several languages for the application. `gotext` returns the first language that provides a translation:

```go
gotext.Configure("locales", "es_UY:es:en_US", "default")

// Falls back es_UY -> es -> en_US in that order
fmt.Println(gotext.Get("Hello, world!"))
```

### Language Simplification

If a file for the full locale code (e.g., `es_UY`) is missing, a file for the base language (e.g., `es`) is used instead. `GetActualLanguage` reports which language code the filesystem actually resolved. Files may live under an `LC_MESSAGES` subdirectory or directly under the language code — both layouts are supported.

### Domain Fallback

Domain-level fallback follows the same idea: when a domain is missing for a given key, the message ID itself is returned, formatted with the provided variables.

## 5. Custom Translation Backends

The `Translator` interface defines what a translation source must provide:

```go
type Translator interface {
    ParseFile(f string)
    Parse(buf []byte)
    Get(str string, vars ...any) string
    GetN(str, plural string, n int, vars ...any) string
    GetC(str, ctx string, vars ...any) string
    GetNC(str, plural string, n int, ctx string, vars ...any) string
    MarshalBinary() ([]byte, error)
    UnmarshalBinary([]byte) error
    GetDomain() *Domain
}
```

You can implement this interface to plug in a different translation backend (e.g., a remote service or a database). Register your implementation on a `Locale` with `AddTranslator(domain, translator)` instead of `AddDomain`:

```go
l := gotext.NewLocale("", "fr_FR")
l.AddTranslator("default", myCustomTranslator)
fmt.Println(l.Get("Hello, world!"))
```

The `AppendTranslator` interface extends `Translator` with `Append`/`AppendN`/`AppendC`/`AppendNC` buffer builders, used to stream serialized output.

### In-Memory PO Manipulation

For tests or programmatic use, `Po` objects can be built entirely in memory:

```go
po := gotext.NewPo()
po.Parse([]byte(`msgid "Hello, world!"
msgstr "Hola, mundo!"`))
fmt.Println(po.Get("Hello, world!"))
```

`Set`, `SetN`, `SetC`, and `SetNC` add or overwrite translations directly, and `MarshalText` produces a PO-format text representation for round-tripping. `DropStaleTranslations`, `SetRefs`, and `GetRefs` help maintain translation files generated by `xgotext`.