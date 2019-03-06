package gotext

/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

// Gettext ...
func (l *Locale) Gettext(msgid string) string {
	return l.Get(msgid)
}

// Dgettext ...
func (l *Locale) Dgettext(domain, msgid string) string {
	return l.GetD(domain, msgid)
}

// Ngettext ...
func (l *Locale) Ngettext(msgid, msgidPlural string, count int) string {
	return l.GetN(msgid, msgidPlural, count)
}

// Dngettext ...
func (l *Locale) Dngettext(domain, msgid, msgidPlural string, count int) string {
	return l.GetND(domain, msgid, msgidPlural, count)
}

// Pgettext ...
func (l *Locale) Pgettext(msgctxt, msgid string) string {
	return l.GetC(msgid, msgctxt)
}

// Dpgettext ...
func (l *Locale) Dpgettext(domain, msgctxt, msgid string) string {
	return l.GetDC(domain, msgid, msgctxt)
}

// Npgettext ...
func (l *Locale) Npgettext(msgctxt, msgid, msgidPlural string, count int) string {
	return l.GetNC(msgid, msgidPlural, count, msgctxt)
}

// Dnpgettext ...
func (l *Locale) Dnpgettext(domain, msgctxt, msgid, msgidPlural string, count int) string {
	return l.GetNDC(domain, msgid, msgidPlural, count, msgctxt)
}
