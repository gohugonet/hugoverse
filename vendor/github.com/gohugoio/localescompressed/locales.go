// Package localescompressed compresses and wraps all translators in github.com/gohugoio/locales.
// The translators are not created until asked for in Get.
package localescompressed

import (
	"strings"
	"sync"

	"github.com/gohugoio/locales"
	"github.com/gohugoio/locales/currency"
)

var (
	// One normally only need a small subset of all the languages,
	// so delay creation until needed.
	mu              sync.RWMutex
	translatorFuncs = make(map[string]func() locales.Translator)
	translators     = make(map[string]locales.Translator)
)

// GetTranslator gets the Translator for the given locale, nil if not found.
func GetTranslator(locale string) locales.Translator {
	locale = strings.ToLower(strings.ReplaceAll(locale, "-", "_"))

	mu.RLock()
	t, found := translators[locale]
	if found {
		mu.RUnlock()
		return t
	}

	fn, found := translatorFuncs[locale]
	mu.RUnlock()
	if !found {
		return nil
	}

	mu.Lock()
	t = fn()
	translators[locale] = t
	mu.Unlock()

	return t
}

// GetCurrency gets the currency for the given ISO 4217 code,
// or -1 if not found.
func GetCurrency(code string) currency.Type {
	c, found := currencies[strings.ToUpper(code)]
	if !found {
		return -1
	}
	return c
}
