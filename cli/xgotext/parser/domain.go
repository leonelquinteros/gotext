package parser

import (
	"sort"
	"strings"
)

// Translation for a text to translate
type Translation struct {
	MsgId           string
	MsgIdPlural     string
	Context         string
	SourceLocations []string
}

// AddLocations to translation
func (t *Translation) AddLocations(locations []string) {
	if t.SourceLocations == nil {
		t.SourceLocations = locations
	} else {
		t.SourceLocations = append(t.SourceLocations, locations...)
	}
}

// Dump translation as string
func (t *Translation) Dump() string {
	data := make([]string, 0, len(t.SourceLocations)+5)

	for _, location := range t.SourceLocations {
		data = append(data, "#: "+location)
	}

	if t.Context != "" {
		data = append(data, "msgctxt "+t.Context)
	}

	data = append(data, "msgid "+t.MsgId)

	if t.MsgIdPlural == "" {
		data = append(data, "msgstr \"\"")
	} else {
		data = append(data,
			"msgid_plural "+t.MsgIdPlural,
			"msgstr[0] \"\"",
			"msgstr[1] \"\"")
	}

	return strings.Join(data, "\n")
}

// TranslationMap contains a map of translations with the ID as key
type TranslationMap map[string]*Translation

// Dump the translation map as string
func (m TranslationMap) Dump() string {
	// sort by translation id for consistence output
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := make([]string, 0, len(m))
	for _, key := range keys {
		data = append(data, (m)[key].Dump())
	}
	return strings.Join(data, "\n\n")
}

// Domain holds all translations of one domain
type Domain struct {
	Translations        TranslationMap
	ContextTranslations map[string]TranslationMap
}

// AddTranslation to the domain
func (d *Domain) AddTranslation(translation *Translation) {
	if d.Translations == nil {
		d.Translations = make(TranslationMap)
		d.ContextTranslations = make(map[string]TranslationMap)
	}

	if translation.Context == "" {
		if t, ok := d.Translations[translation.MsgId]; ok {
			t.AddLocations(translation.SourceLocations)
		} else {
			d.Translations[translation.MsgId] = translation
		}
	} else {
		if _, ok := d.ContextTranslations[translation.Context]; !ok {
			d.ContextTranslations[translation.Context] = make(TranslationMap)
		}

		if t, ok := d.ContextTranslations[translation.Context][translation.MsgId]; ok {
			t.AddLocations(translation.SourceLocations)
		} else {
			d.ContextTranslations[translation.Context][translation.MsgId] = translation
		}
	}
}

// Dump the domain as string
func (d *Domain) Dump() string {
	data := make([]string, 0, len(d.ContextTranslations)+1)
	data = append(data, d.Translations.Dump())

	// sort context translations by context for consistence output
	keys := make([]string, 0, len(d.ContextTranslations))
	for k := range d.ContextTranslations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		data = append(data, d.ContextTranslations[key].Dump())
	}
	return strings.Join(data, "\n\n")
}

// DomainMap contains multiple domains as map with name as key
type DomainMap map[string]*Domain

// AddTranslation to domain map
func (m *DomainMap) AddTranslation(domain string, translation *Translation) {
	if _, ok := (*m)[domain]; !ok {
		(*m)[domain] = new(Domain)
	}
	(*m)[domain].AddTranslation(translation)
}
