package gotext

import (
	"fmt"
	"os"
	"path"
)

type Locale struct {
	// Path to locale files.
	path string

	// Language for this Locale
	lang string

	// List of available domains for this locale.
	domains map[string]*Po
}

// NewLocale creates and initializes a new Locale object for a given language.
// It receives a path for the i18n files directory (p) and a language code to use (l).
func NewLocale(p, l string) *Locale {
	return &Locale{
		path:    p,
		lang:    l,
		domains: make(map[string]*Po),
	}
}

// AddDomain creates a new domain for a given locale object and initializes the Po object.
// If the domain exists, it gets reloaded.
func (l *Locale) AddDomain(dom string) {
	po := new(Po)

	// Check for file.
	filename := path.Clean(l.path + string(os.PathSeparator) + l.lang + string(os.PathSeparator) + dom + ".po")

	// Try to use the generic language dir if the provided isn't available
	if _, err := os.Stat(filename); err != nil {
		if len(l.lang) > 2 {
			filename = path.Clean(l.path + string(os.PathSeparator) + l.lang[:2] + string(os.PathSeparator) + dom + ".po")
		}
	}

	// Parse file.
	po.ParseFile(filename)

	// Save new domain
	if l.domains == nil {
		l.domains = make(map[string]*Po)
	}
	l.domains[dom] = po
}

// Get uses a domain "default" to return the corresponding translation of a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) Get(str string, vars ...interface{}) string {
	return l.GetD("default", str, vars...)
}

// GetN retrieves the (N)th plural form translation for the given string in the "default" domain.
// If n == 0, usually the singular form of the string is returned as defined in the PO file.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetN(str, plural string, n int, vars ...interface{}) string {
	return l.GetND("default", str, plural, n, vars...)
}

// GetD returns the corresponding translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetD(dom, str string, vars ...interface{}) string {
	return l.GetND(dom, str, str, 0, vars...)
}

// GetND retrieves the (N)th plural form translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (l *Locale) GetND(dom, str, plural string, n int, vars ...interface{}) string {
	if l.domains != nil {
		if _, ok := l.domains[dom]; ok {
			if l.domains[dom] != nil {
				return l.domains[dom].GetN(str, plural, n, vars...)
			}
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(plural, vars...)
}
