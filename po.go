package gotext

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Po struct {
    // Storage
	translations map[string]string
	
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
    po.Lock()
	defer po.Unlock()
	
	if po.translations == nil {
		po.translations = make(map[string]string)
	}

	lines := strings.Split(str, "\n")

	var msgid, msgstr string

	for _, l := range lines {
		// Trim spaces
		l = strings.TrimSpace(l)

		// Skip empty lines
		if l == "" {
			continue
		}

		// Skip invalid lines
		if !strings.HasPrefix(l, "msgid") && !strings.HasPrefix(l, "msgstr") {
			continue
		}

		// Buffer msgid and continue
		if strings.HasPrefix(l, "msgid") {
			msgid = strings.TrimSpace(strings.TrimPrefix(l, "msgid"))
			msgid, _ = strconv.Unquote(msgid)

			continue
		}

		// Save translation for buffered msgid
		if strings.HasPrefix(l, "msgstr") {
			msgstr = strings.TrimSpace(strings.TrimPrefix(l, "msgstr"))
			msgstr, _ = strconv.Unquote(msgstr)

			po.translations[msgid] = msgstr
		}
	}
}

func (po *Po) Get(str string, vars ...interface{}) string {
	if po.translations != nil {
		if _, ok := po.translations[str]; ok {
			return fmt.Sprintf(po.translations[str], vars...)
		}
	}

	// Return the same we received by default
	return fmt.Sprintf(str, vars...)
}
