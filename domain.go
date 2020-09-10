package gotext

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"net/textproto"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/text/language"

	"github.com/leonelquinteros/gotext/plurals"
)

// Domain has all the common functions for dealing with a gettext domain
// it's initialized with a GettextFile (which represents either a Po or Mo file)
type Domain struct {
	Headers textproto.MIMEHeader

	// Language header
	Language string
	tag      language.Tag

	// Plural-Forms header
	PluralForms string

	// Parsed Plural-Forms header values
	nplurals    int
	plural      string
	pluralforms plurals.Expression

	// Storage
	translations       map[string]*Translation
	contexts           map[string]map[string]*Translation
	pluralTranslations map[string]*Translation

	// Sync Mutex
	trMutex     sync.RWMutex
	pluralMutex sync.RWMutex

	// Parsing buffers
	trBuffer  *Translation
	ctxBuffer string
}

func NewDomain() *Domain {
	domain := new(Domain)

	domain.translations = make(map[string]*Translation)
	domain.contexts = make(map[string]map[string]*Translation)
	domain.pluralTranslations = make(map[string]*Translation)

	return domain
}

func (do *Domain) pluralForm(n int) int {
	// do we really need locking here? not sure how this plurals.Expression works, so sticking with it for now
	do.pluralMutex.RLock()
	defer do.pluralMutex.RUnlock()

	// Failure fallback
	if do.pluralforms == nil {
		/* Use the Germanic plural rule.  */
		if n == 1 {
			return 0
		}
		return 1
	}
	return do.pluralforms.Eval(uint32(n))
}

// parseHeaders retrieves data from previously parsed headers. it's called by both Mo and Po when parsing
func (do *Domain) parseHeaders() {
	// Make sure we end with 2 carriage returns.
	empty := ""
	if _, ok := do.translations[empty]; ok {
		empty = do.translations[empty].Get()
	}
	raw := empty + "\n\n"

	// Read
	reader := bufio.NewReader(strings.NewReader(raw))
	tp := textproto.NewReader(reader)

	var err error

	do.Headers, err = tp.ReadMIMEHeader()
	if err != nil {
		return
	}

	// Get/save needed headers
	do.Language = do.Headers.Get("Language")
	do.tag = language.Make(do.Language)
	do.PluralForms = do.Headers.Get("Plural-Forms")

	// Parse Plural-Forms formula
	if do.PluralForms == "" {
		return
	}

	// Split plural form header value
	pfs := strings.Split(do.PluralForms, ";")

	// Parse values
	for _, i := range pfs {
		vs := strings.SplitN(i, "=", 2)
		if len(vs) != 2 {
			continue
		}

		switch strings.TrimSpace(vs[0]) {
		case "nplurals":
			do.nplurals, _ = strconv.Atoi(vs[1])

		case "plural":
			do.plural = vs[1]

			if expr, err := plurals.Compile(do.plural); err == nil {
				do.pluralforms = expr
			}

		}
	}
}

func (do *Domain) Get(str string, vars ...interface{}) string {
	// Sync read
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.translations != nil {
		if _, ok := do.translations[str]; ok {
			return Printf(do.translations[str].Get(), vars...)
		}
	}

	// Return the same we received by default
	return Printf(str, vars...)
}

// GetN retrieves the (N)th plural form of Translation for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (do *Domain) GetN(str, plural string, n int, vars ...interface{}) string {
	// Sync read
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.translations != nil {
		if _, ok := do.translations[str]; ok {
			return Printf(do.translations[str].GetN(do.pluralForm(n)), vars...)
		}
	}

	// Parse plural forms to distinguish between plural and singular
	if do.pluralForm(n) == 0 {
		return Printf(str, vars...)
	}
	return Printf(plural, vars...)
}

// GetC retrieves the corresponding Translation for a given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (do *Domain) GetC(str, ctx string, vars ...interface{}) string {
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.contexts != nil {
		if _, ok := do.contexts[ctx]; ok {
			if do.contexts[ctx] != nil {
				if _, ok := do.contexts[ctx][str]; ok {
					return Printf(do.contexts[ctx][str].Get(), vars...)
				}
			}
		}
	}

	// Return the string we received by default
	return Printf(str, vars...)
}

// GetNC retrieves the (N)th plural form of Translation for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (do *Domain) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.contexts != nil {
		if _, ok := do.contexts[ctx]; ok {
			if do.contexts[ctx] != nil {
				if _, ok := do.contexts[ctx][str]; ok {
					return Printf(do.contexts[ctx][str].GetN(do.pluralForm(n)), vars...)
				}
			}
		}
	}

	if n == 1 {
		return Printf(str, vars...)
	}
	return Printf(plural, vars...)
}

// MarshalBinary implements encoding.BinaryMarshaler interface
func (do *Domain) MarshalBinary() ([]byte, error) {
	obj := new(TranslatorEncoding)
	obj.Headers = do.Headers
	obj.Language = do.Language
	obj.PluralForms = do.PluralForms
	obj.Nplurals = do.nplurals
	obj.Plural = do.plural
	obj.Translations = do.translations
	obj.Contexts = do.contexts

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(obj)

	return buff.Bytes(), err
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface
func (do *Domain) UnmarshalBinary(data []byte) error {
	buff := bytes.NewBuffer(data)
	obj := new(TranslatorEncoding)

	decoder := gob.NewDecoder(buff)
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	do.Headers = obj.Headers
	do.Language = obj.Language
	do.PluralForms = obj.PluralForms
	do.nplurals = obj.Nplurals
	do.plural = obj.Plural
	do.translations = obj.Translations
	do.contexts = obj.Contexts

	if expr, err := plurals.Compile(do.plural); err == nil {
		do.pluralforms = expr
	}

	return nil
}
