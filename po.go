package gotext

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

type translation struct {
	id       string
	pluralId string
	trs      map[int]string
}

func newTranslation() *translation {
	tr := new(translation)
	tr.trs = make(map[int]string)

	return tr
}

func (t *translation) get() string {
	// Look for translation index 0
	if _, ok := t.trs[0]; ok {
		return t.trs[0]
	}

	// Return unstranlated id by default
	return t.id
}

func (t *translation) getN(n int) string {
	// Look for translation index
	if _, ok := t.trs[n]; ok {
		return t.trs[n]
	}

	// Return unstranlated plural by default
	return t.pluralId
}

/*
Po parses the content of any PO file and provides all the translation functions needed.

Example:

	import "github.com/leonelquinteros/gotext"

    func main() {
    	// Create po object
        po := new(Po)

        // Parse .po file
        po.ParseFile("/path/to/po/file/translations.po")

        // Get translation
        println(po.Get("Translate this"))
    }
*/
type Po struct {
	// Storage
	translations map[string]*translation

	// Sync Mutex
	sync.RWMutex
}

// ParseFile tries to read the file by its provided path (f) and parse its content as a .po file.
func (po *Po) ParseFile(f string) {
	// Check if file exists
	info, err := os.Stat(f)
	if err != nil {
		return
	}

	// Check that isn't a directory
	if info.IsDir() {
		return
	}

	// Parse file content
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}

	po.Parse(string(data))
}

// Parse loads the translations specified in the provided string (str)
func (po *Po) Parse(str string) {
	if po.translations == nil {
		po.Lock()
		po.translations = make(map[string]*translation)
		po.Unlock()
	}

	lines := strings.Split(str, "\n")

	tr := newTranslation()

	for _, l := range lines {
		// Trim spaces
		l = strings.TrimSpace(l)

		// Skip empty lines
		if l == "" {
			continue
		}

		// Skip invalid lines
		if !strings.HasPrefix(l, "msgid") && !strings.HasPrefix(l, "msgid_plural") && !strings.HasPrefix(l, "msgstr") {
			continue
		}

		// Buffer msgid and continue
		if strings.HasPrefix(l, "msgid") && !strings.HasPrefix(l, "msgid_plural") {
			// Save current translation buffer.
			po.Lock()
			po.translations[tr.id] = tr
			po.Unlock()

			// Flush buffer
			tr = newTranslation()

			// Set id
			tr.id, _ = strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(l, "msgid")))

			// Loop
			continue
		}

		// Check for plural form
		if strings.HasPrefix(l, "msgid_plural") {
			tr.pluralId, _ = strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(l, "msgid_plural")))

			// Loop
			continue
		}

		// Save translation
		if strings.HasPrefix(l, "msgstr") {
			l = strings.TrimSpace(strings.TrimPrefix(l, "msgstr"))

			// Check for indexed translation forms
			if strings.HasPrefix(l, "[") {
				in := strings.Index(l, "]")
				if in == -1 {
					// Skip wrong index formatting
					continue
				}

				// Parse index
				i, err := strconv.Atoi(l[1:in])
				if err != nil {
					// Skip wrong index formatting
					continue
				}

				// Parse translation string
				tr.trs[i], _ = strconv.Unquote(strings.TrimSpace(l[in+1:]))

				// Loop
				continue
			}

			// Save single translation form under 0 index
			tr.trs[0], _ = strconv.Unquote(l)
		}
	}

	// Save last translation buffer.
	if tr.id != "" {
		po.Lock()
		po.translations[tr.id] = tr
		po.Unlock()
	}
}

// Get retrieves the corresponding translation for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) Get(str string, vars ...interface{}) string {
	if po.translations != nil {
		if _, ok := po.translations[str]; ok {
			return fmt.Sprintf(po.translations[str].get(), vars...)
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(str, vars...)
}

// GetN retrieves the (N)th plural form translation for the given string.
// If n == 0, usually the singular form of the string is returned as defined in the PO file.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (po *Po) GetN(str, plural string, n int, vars ...interface{}) string {
	if po.translations != nil {
		if _, ok := po.translations[str]; ok {
			return fmt.Sprintf(po.translations[str].getN(n), vars...)
		}
	}

	// Return the plural string we received by default
	return fmt.Sprintf(plural, vars...)
}
