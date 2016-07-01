[![GoDoc](https://godoc.org/github.com/leonelquinteros/gotext?status.svg)](https://godoc.org/github.com/leonelquinteros/gotext)

# Gotext

GNU gettext utilities for Go. 

Version: [v1.0.1](https://github.com/leonelquinteros/gotext/releases/tag/v1.0.1)


# Features

- Implements GNU gettext support in native Go.
- Complete support for [PO files](https://www.gnu.org/software/gettext/manual/html_node/PO-Files.html).
- Support for [pluralization rules](https://www.gnu.org/software/gettext/manual/html_node/Plural-forms.html).
- Support for [message context](https://www.gnu.org/software/gettext/manual/html_node/Contexts.html).
- Support for variables inside translation strings using Go's [fmt package syntax](https://golang.org/pkg/fmt/).
- Thread-safe: This package is safe for concurrent use across multiple goroutines. 
- It works with UTF-8 encoding as it's the default for Go language.
- Unit tests available.
- Language codes are automatically simplified from the form `en_UK` to `en` if the first isn't available.
- Ready to use inside Go templates.


# License

[MIT license](LICENSE)


# Documentation

Refer to the Godoc package documentation at (https://godoc.org/github.com/leonelquinteros/gotext)


# Installation 

```
go get github.com/leonelquinteros/gotext
```

- There are no requirements or dependencies to use this package. 
- No need to install GNU gettext utilities (unless specific needs of CLI tools).
- No need for environment variables. Some naming conventions are applied but not needed.  


#### Version vendoring

Stable releases use [semantic versioning](http://semver.org/spec/v2.0.0.html) tagging on this repository.

You can rely on this to use your preferred vendoring tool or to manually retrieve the corresponding release tag from the GitHub repository.


##### Vendoring with [gopkg.in](http://labix.org/gopkg.in)

[http://gopkg.in/leonelquinteros/gotext.v1](http://gopkg.in/leonelquinteros/gotext.v1)

To get the latest v1 package stable release, execute:

```
go get gopkg.in/leonelquinteros/gotext.v1
```

To import this package, add the following line to your code:

```go
import "gopkg.in/leonelquinteros/gotext.v1"
```

Refer to it as gotext.


# Locales directories structure

The package will assume a directories structure starting with a base path that will be provided to the package configuration 
or to object constructors depending on the use, but either will use the same convention to lookup inside the base path. 

Inside the base directory where will be the language directories named using the language and country 2-letter codes (en_US, es_AR, ...). 
All package functions can lookup after the simplified version for each language in case the full code isn't present but the more general language code exists. 
So if the language set is `en_UK`, but there is no directory named after that code and there is a directory named `en`, 
all package functions will be able to resolve this generalization and provide translations for the more general library.  

The language codes are assumed to be [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) codes (2-letter codes). 
That said, most functions will work with any coding standard as long the directory name matches the language code set on the configuration.

Then, there can be a `LC_MESSAGES` containing all PO files or the PO files themselves.  
A library directory structure can look like: 

```
/path/to/locales
/path/to/locales/en_US
/path/to/locales/en_US/LC_MESSAGES
/path/to/locales/en_US/LC_MESSAGES/default.po
/path/to/locales/en_US/LC_MESSAGES/extras.po
/path/to/locales/en_UK
/path/to/locales/en_UK/LC_MESSAGES
/path/to/locales/en_UK/LC_MESSAGES/default.po
/path/to/locales/en_UK/LC_MESSAGES/extras.po
/path/to/locales/en_AU
/path/to/locales/en_AU/LC_MESSAGES
/path/to/locales/en_AU/LC_MESSAGES/default.po
/path/to/locales/en_AU/LC_MESSAGES/extras.po
/path/to/locales/es
/path/to/locales/es/default.po
/path/to/locales/es/extras.po
/path/to/locales/es_ES
/path/to/locales/es_ES/default.po
/path/to/locales/es_ES/extras.po
/path/to/locales/fr
/path/to/locales/fr/default.po
/path/to/locales/fr/extras.po
``` 

And so on...



# About translation function names

The standard GNU gettext defines helper functions that maps to the `gettext()` function and it's widely adopted by most implementations. 

The basic translation function is usually `_()` in the form: 

```
_("Translate this")
``` 

In Go, this can't be implemented by a reusable package as the function name has to start with a capital letter in order to be exported. 

Each implementation of this package can declare this helper functions inside their own packages if this function naming are desired/needed: 

```go
package main

import "github.com/leonelquinteros/gotext"

func _(str string, vars ...interface{}) string {
    return gotext.Get(str, vars...)
}

``` 

This is valid and can be used within a package.

In normal conditions the Go compiler will optimize the calls to `_()` by replacing its content in place of the function call to reduce the function calling overhead. 
This is a normal Go compiler behavior.  



# Usage examples

## Using package for single language/domain settings

For quick/simple translations you can use the package level functions directly.

```go
import "github.com/leonelquinteros/gotext"

func main() {
    // Configure package
    gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")
    
    // Translate text from default domain
    println(gotext.Get("My text on 'domain-name' domain"))
    
    // Translate text from a different domain without reconfigure
    println(gotext.GetD("domain2", "Another text on a different domain"))
}

```

## Using dynamic variables on translations

All translation strings support dynamic variables to be inserted without translate. 
Use the fmt.Printf syntax (from Go's "fmt" package) to specify how to print the non-translated variable inside the translation string. 

```go
import "github.com/leonelquinteros/gotext"

func main() {
    // Configure package
    gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")
    
    // Set variables
    name := "John"
    
    // Translate text with variables
    println(gotext.Get("Hi, my name is %s", name))
}

```


## Using Locale object

When having multiple languages/domains/libraries at the same time, you can create Locale objects for each variation 
so you can handle each settings on their own.

```go
import "github.com/leonelquinteros/gotext"

func main() {
    // Create Locale with library path and language code
    l := gotext.NewLocale("/path/to/locales/root/dir", "es_UY")
    
    // Load domain '/path/to/locales/root/dir/es_UY/default.po'
    l.AddDomain("default")
    
    // Translate text from default domain
    println(l.Get("Translate this"))
    
    // Load different domain
    l.AddDomain("translations")
    
    // Translate text from domain
    println(l.GetD("translations", "Translate this"))
}
```

This is also helpful for using inside templates (from the "text/template" package), where you can pass the Locale object to the template.
If you set the Locale object as "Loc" in the template, then the template code would look like: 

```
{{ .Loc.Get "Translate this" }}
```


## Using the Po object to handle .po files and PO-formatted strings

For when you need to work with PO files and strings, 
you can directly use the Po object to parse it and access the translations in there in the same way.

```go
import "github.com/leonelquinteros/gotext"

func main() {
    // Set PO content
    str := `
msgid "Translate this"
msgstr "Translated text"

msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgstr "This one sets the var: %s"
`
    
    // Create Po object
    po := new(Po)
    po.Parse(str)
    
    println(po.Get("Translate this"))
}
```


## Use plural forms of translations

PO format supports defining one or more plural forms for the same translation.

```go
import "github.com/leonelquinteros/gotext"

func main() {
    // Set PO content
    str := `
msgid "Translate this"
msgstr "Translated text"

msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular: %s"
msgstr[1] "This one is the plural: %s"
msgstr[2] "And this is the second plural form: %s"
`
    
    // Create Po object
    po := new(Po)
    po.Parse(str)
    
    println(po.GetN("One with var: %s", "Several with vars: %s", 2, v))
    // "And this is the second plural form: Variable"
}
```


# Contribute 

- Please, contribute.
- Use the package on your projects.
- Report issues on Github. 
- Send pull requests for bugfixes and improvements.
- Send proposals on Github issues.
