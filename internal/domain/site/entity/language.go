package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/site"
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
	return l.currentLanguage
}

func (l *Language) Lang() string {
	return l.currentLanguage
}

func (l *Language) LanguageCode() string {
	return l.currentLanguage
}

func (l *Language) LanguageDirection() string {
	return "ltr"
}
