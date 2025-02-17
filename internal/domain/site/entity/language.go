package entity

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/site"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"time"
)

type Language struct {
	LangSvc site.LanguageService

	currentLocation *time.Location
	currentLanguage string
	collator        *Collator
}

func (l *Language) CurrentLanguageIndex() int {
	curr, err := l.LangSvc.GetLanguageIndex(l.currentLanguage)
	if err != nil {
		panic(fmt.Sprintf("language %q not found", l.currentLanguage))
	}
	return curr
}

func (l *Language) setup() error {
	// TODO: make it configurable from config timeZone field
	l.currentLocation = time.UTC
	return nil
}

func (l *Language) Languages() []Language {
	var langs []Language

	for _, lang := range l.LangSvc.LanguageKeys() {
		langs = append(langs, Language{
			LangSvc:         l.LangSvc,
			currentLocation: l.currentLocation,
			currentLanguage: lang,
			collator:        l.collator,
		})
	}

	return langs
}

func (l *Language) Collator() *collate.Collator {
	if l.collator == nil {
		tag, err := language.Parse(l.currentLanguage)
		if err == nil {
			l.collator = &Collator{
				c: collate.New(tag),
			}
		} else {
			l.collator = &Collator{
				c: collate.New(language.English),
			}
		}
	}

	return l.collator.c
}

func (l *Language) Location() *time.Location {
	return l.currentLocation
}

func (l *Language) isMultipleLanguage() bool {
	return len(l.LangSvc.LanguageKeys()) > 1
}

func (l *Language) LanguagePrefix() string {
	if l.currentLanguage == l.LangSvc.DefaultLanguage() {
		return ""
	}
	return l.currentLanguage
}

func (l *Language) Lang() string {
	return l.currentLanguage
}

func (l *Language) LanguageName() string {
	return l.LangSvc.GetLanguageName(l.currentLanguage)
}

func (l *Language) DefaultLanguageName() string {
	return l.LangSvc.GetLanguageName(l.LangSvc.DefaultLanguage())
}

func (l *Language) LanguageCode() string {
	return l.currentLanguage
}

func (l *Language) LanguageDirection() string {
	return "ltr"
}
