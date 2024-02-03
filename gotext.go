/*
Package gotext implements GNU gettext utilities.

For quick/simple translations you can use the package level functions directly.

	    import (
		    "fmt"
		    "github.com/donseba/gotext"
	    )

	    func main() {
	        // Configure package
	        gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

	        // Translate text from default domain
	        fmt.Println(gotext.Get("My text on 'domain-name' domain"))

	        // Translate text from a different domain without reconfigure
	        fmt.Println(gotext.GetD("domain2", "Another text on a different domain"))
	    }
*/
package gotext

import (
	"encoding/gob"
	"sync"
)

// Global environment variables
type config struct {
	sync.RWMutex

	// Default domain to look at when no domain is specified. Used by package level functions.
	domain string

	// Language set.
	language string

	// Path to library directory where all locale directories and Translation files are.
	library string

	// Storage for package level methods
	storage *Locale
}

var globalConfig *config

func init() {
	// Init default configuration
	globalConfig = &config{
		domain:   "default",
		language: "en_US",
		library:  "/usr/local/share/locale",
		storage:  nil,
	}

	// Register Translator types for gob encoding
	gob.Register(TranslatorEncoding{})
}

// loadStorage creates a new Locale object at package level based on the Global variables settings.
// It's called automatically when trying to use Get or GetD methods.
func loadStorage(force bool) {
	globalConfig.Lock()

	if globalConfig.storage == nil || force {
		globalConfig.storage = NewLocale(globalConfig.library, globalConfig.language)
	}

	if _, ok := globalConfig.storage.Domains[globalConfig.domain]; !ok || force {
		globalConfig.storage.AddDomain(globalConfig.domain)
	}
	globalConfig.storage.SetDomain(globalConfig.domain)

	globalConfig.Unlock()
}

// GetDomain is the domain getter for the package configuration
func GetDomain() string {
	var dom string
	globalConfig.RLock()
	if globalConfig.storage != nil {
		dom = globalConfig.storage.GetDomain()
	}
	if dom == "" {
		dom = globalConfig.domain
	}
	globalConfig.RUnlock()

	return dom
}

// SetDomain sets the name for the domain to be used at package level.
// It reloads the corresponding Translation file.
func SetDomain(dom string) {
	globalConfig.Lock()
	globalConfig.domain = dom
	if globalConfig.storage != nil {
		globalConfig.storage.SetDomain(dom)
	}
	globalConfig.Unlock()

	loadStorage(true)
}

// GetLanguage is the language getter for the package configuration
func GetLanguage() string {
	globalConfig.RLock()
	lang := globalConfig.language
	globalConfig.RUnlock()

	return lang
}

// SetLanguage sets the language code to be used at package level.
// It reloads the corresponding Translation file.
func SetLanguage(lang string) {
	globalConfig.Lock()
	globalConfig.language = SimplifiedLocale(lang)
	globalConfig.Unlock()

	loadStorage(true)
}

// GetLibrary is the library getter for the package configuration
func GetLibrary() string {
	globalConfig.RLock()
	lib := globalConfig.library
	globalConfig.RUnlock()

	return lib
}

// SetLibrary sets the root path for the loale directories and files to be used at package level.
// It reloads the corresponding Translation file.
func SetLibrary(lib string) {
	globalConfig.Lock()
	globalConfig.library = lib
	globalConfig.Unlock()

	loadStorage(true)
}

// GetStorage is the locale storage getter for the package configuration.
func GetStorage() *Locale {
	globalConfig.RLock()
	storage := globalConfig.storage
	globalConfig.RUnlock()

	return storage
}

// SetStorage allows overridding the global Locale object with one built manually with NewLocale().
// This allows then to attach to the locale Domains object in memory po or mo files (embedded or in any directory),
// for each domain.
// Locale library, language and domain properties will apply on default global configuration.
// Any domain not loaded yet will use to the just in time domain loading process.
// Note that any call to gotext.Set* or Configure will invalidate this override.
func SetStorage(storage *Locale) {
	globalConfig.Lock()
	globalConfig.storage = storage
	globalConfig.library = storage.path
	globalConfig.language = storage.lang
	globalConfig.domain = storage.defaultDomain
	globalConfig.Unlock()
}

// Configure sets all configuration variables to be used at package level and reloads the corresponding Translation file.
// It receives the library path, language code and domain name.
// This function is recommended to be used when changing more than one setting,
// as using each setter will introduce a I/O overhead because the Translation file will be loaded after each set.
func Configure(lib, lang, dom string) {
	globalConfig.Lock()
	globalConfig.library = lib
	globalConfig.language = SimplifiedLocale(lang)
	globalConfig.domain = dom
	globalConfig.Unlock()

	loadStorage(true)
}

// Get uses the default domain globally set to return the corresponding Translation of a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func Get(str string, vars ...interface{}) string {
	return GetD(GetDomain(), str, vars...)
}

// GetN retrieves the (N)th plural form of Translation for the given string in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetN(str, plural string, n int, vars ...interface{}) string {
	return GetND(GetDomain(), str, plural, n, vars...)
}

// GetD returns the corresponding Translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetD(dom, str string, vars ...interface{}) string {
	// Try to load default package Locale storage
	loadStorage(false)

	// Return Translation
	globalConfig.RLock()

	if _, ok := globalConfig.storage.Domains[dom]; !ok {
		globalConfig.storage.AddDomain(dom)
	}

	tr := globalConfig.storage.GetD(dom, str, vars...)
	globalConfig.RUnlock()

	return tr
}

// GetND retrieves the (N)th plural form of Translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetND(dom, str, plural string, n int, vars ...interface{}) string {
	// Try to load default package Locale storage
	loadStorage(false)

	// Return Translation
	globalConfig.RLock()

	if _, ok := globalConfig.storage.Domains[dom]; !ok {
		globalConfig.storage.AddDomain(dom)
	}

	tr := globalConfig.storage.GetND(dom, str, plural, n, vars...)
	globalConfig.RUnlock()

	return tr
}

// GetC uses the default domain globally set to return the corresponding Translation of the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetC(str, ctx string, vars ...interface{}) string {
	return GetDC(GetDomain(), str, ctx, vars...)
}

// GetNC retrieves the (N)th plural form of Translation for the given string in the given context in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return GetNDC(GetDomain(), str, plural, n, ctx, vars...)
}

// GetDC returns the corresponding Translation in the given domain for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetDC(dom, str, ctx string, vars ...interface{}) string {
	// Try to load default package Locale storage
	loadStorage(false)

	// Return Translation
	globalConfig.RLock()
	tr := globalConfig.storage.GetDC(dom, str, ctx, vars...)
	globalConfig.RUnlock()

	return tr
}

// GetNDC retrieves the (N)th plural form of Translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetNDC(dom, str, plural string, n int, ctx string, vars ...interface{}) string {
	// Try to load default package Locale storage
	loadStorage(false)

	// Return Translation
	globalConfig.RLock()
	tr := globalConfig.storage.GetNDC(dom, str, plural, n, ctx, vars...)
	globalConfig.RUnlock()

	return tr
}

// IsTranslated reports whether a string is translated
func IsTranslated(str string) bool {
	return IsTranslatedND(GetDomain(), str, 1)
}

// IsTranslatedN reports whether a plural string is translated
func IsTranslatedN(str string, n int) bool {
	return IsTranslatedND(GetDomain(), str, n)
}

// IsTranslatedD reports whether a domain string is translated
func IsTranslatedD(dom, str string) bool {
	return IsTranslatedND(dom, str, 1)
}

// IsTranslatedND reports whether a plural domain string is translated
func IsTranslatedND(dom, str string, n int) bool {
	loadStorage(false)

	globalConfig.RLock()
	defer globalConfig.RUnlock()

	if _, ok := globalConfig.storage.Domains[dom]; !ok {
		globalConfig.storage.AddDomain(dom)
	}

	return globalConfig.storage.IsTranslatedND(dom, str, n)
}

// IsTranslatedC reports whether a context string is translated
func IsTranslatedC(str, ctx string) bool {
	return IsTranslatedNDC(GetDomain(), str, 1, ctx)
}

// IsTranslatedNC reports whether a plural context string is translated
func IsTranslatedNC(str string, n int, ctx string) bool {
	return IsTranslatedNDC(GetDomain(), str, n, ctx)
}

// IsTranslatedDC reports whether a domain context string is translated
func IsTranslatedDC(dom, str, ctx string) bool {
	return IsTranslatedNDC(dom, str, 1, ctx)
}

// IsTranslatedNDC reports whether a plural domain context string is translated
func IsTranslatedNDC(dom, str string, n int, ctx string) bool {
	loadStorage(false)

	globalConfig.RLock()
	defer globalConfig.RUnlock()

	return globalConfig.storage.IsTranslatedNDC(dom, str, n, ctx)
}
