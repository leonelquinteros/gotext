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
func (l *Locale) Get(str string, vars ...interface{}) string {
	return l.GetD("default", str, vars...)
}

// GetD returns the corresponding translation in the given domain for a given string.
func (l *Locale) GetD(dom, str string, vars ...interface{}) string {
	if l.domains != nil {
		if _, ok := l.domains[dom]; ok {
			if l.domains[dom] != nil {
				return l.domains[dom].Get(str, vars...)
			}
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(str, vars...)
}
