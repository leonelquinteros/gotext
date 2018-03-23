/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

type Translation struct {
	ID       string
	PluralID string
	Trs      map[int]string
}

func NewTranslation() *Translation {
	tr := new(Translation)
	tr.Trs = make(map[int]string)

	return tr
}

func (t *Translation) Get() string {
	// Look for Translation index 0
	if _, ok := t.Trs[0]; ok {
		if t.Trs[0] != "" {
			return t.Trs[0]
		}
	}

	// Return untranslated id by default
	return t.ID
}

func (t *Translation) GetN(n int) string {
	// Look for Translation index
	if _, ok := t.Trs[n]; ok {
		if t.Trs[n] != "" {
			return t.Trs[n]
		}
	}

	// Return untranslated singular if corresponding
	if n == 0 {
		return t.ID
	}

	// Return untranslated plural by default
	return t.PluralID
}
