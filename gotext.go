/*
Package gotext implements GNU gettext utilities.
*/
package gotext

// Global environment variables
var (
	// Default domain to look at when no domain is specified. Used by package level functions.
	domain = "default"

	// Language set.
	language = "en_US"

	// Path to library directory where all locale directories and translation files are.
	library = "/tmp"

	// Storage for package level methods
	storage *Locale
)

// loadStorage creates a new Locale object at package level based on the Global variables settings.
// It's called automatically when trying to use Get or GetD methods.
func loadStorage(force bool) {
	if storage == nil || force {
		storage = NewLocale(library, language)
	}

	if _, ok := storage.domains[domain]; !ok || force {
		storage.AddDomain(domain)
	}
}

// GetDomain is the domain getter for the package configuration
func GetDomain() string {
	return domain
}

// SetDomain sets the name for the domain to be used at package level.
// It reloads the corresponding translation file.
func SetDomain(dom string) {
	domain = dom
	loadStorage(true)
}

// GetLanguage is the language getter for the package configuration
func GetLanguage() string {
	return language
}

// SetLanguage sets the language code to be used at package level.
// It reloads the corresponding translation file.
func SetLanguage(lang string) {
	language = lang
	loadStorage(true)
}

// GetLibrary is the library getter for the package configuration
func GetLibrary() string {
	return library
}

// SetLibrary sets the root path for the loale directories and files to be used at package level.
// It reloads the corresponding translation file.
func SetLibrary(lib string) {
	library = lib
	loadStorage(true)
}

// Configure sets all configuration variables to be used at package level and reloads the corresponding translation file.
// It receives the library path, language code and domain name.
// This function is recommended to be used when changing more than one setting,
// as using each setter will introduce a I/O overhead because the translation file will be loaded after each set.
func Configure(lib, lang, dom string) {
	library = lib
	language = lang
	domain = dom

	loadStorage(true)
}

// Get uses the default domain globally set to return the corresponding translation of a given string.
func Get(str string, vars ...interface{}) string {
	return GetD(domain, str, vars...)
}

// GetD returns the corresponding translation in the given domain for a given string.
func GetD(dom, str string, vars ...interface{}) string {
	// Try to load default package Locale storage
	loadStorage(false)

	// Return translation
	return storage.GetD(dom, str, vars...)
}
