package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"time"
)

type Language struct {
	Config []site.LanguageConfig

	currentLocation *time.Location
	currentLanguage site.LanguageConfig
	collator        *Collator
}

func (l *Language) setup() error {
	// TODO: make it configurable from config timeZone field
	l.currentLocation = time.UTC
	return nil
}

func (l *Language) Collator() *collate.Collator {
	if l.collator == nil {
		tag, err := language.Parse(l.currentLanguage.Code())
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
	return len(l.Config) > 1
}

func (l *Language) DefaultContentLanguage() site.LanguageConfig {
	return l.Config[0]
}

func (l *Language) LanguagePrefix() string {
	if !l.isMultipleLanguage() || l.DefaultContentLanguage().Code() == l.currentLanguage.Code() {
		return ""
	}
	return l.currentLanguage.Code()
}
