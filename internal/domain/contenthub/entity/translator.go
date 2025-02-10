package entity

import (
	"context"
	"errors"
	"fmt"
	"github.com/gohugoio/go-i18n/v2/i18n"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/hreflect"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

const ArtificialLangTagPrefix = "art-x-"

type TranslateFunc func(ctx context.Context, translationID string, templateData any) string

type Translator struct {
	ContentLanguage string
	TranslateFuncs  map[string]TranslateFunc

	Log loggers.Logger `json:"-"`
}

func (t *Translator) Translate(ctx context.Context, lang string, translationID string, templateData any) string {
	if f, ok := t.TranslateFuncs[lang]; ok {
		return f(ctx, translationID, templateData)
	}

	t.Log.Infof("Translation func for language %v not found, use default.", lang)
	if f, ok := t.TranslateFuncs[t.ContentLanguage]; ok {
		return f(ctx, translationID, templateData)
	}

	t.Log.Infoln("i18n not initialized; if you need string translations, check that you have a bundle in /i18n that matches the site language or the default language.")

	return ""
}

func (t *Translator) SetupTranslateFuncs(bndl *i18n.Bundle) {
	enableMissingTranslationPlaceholders := true

	for _, lang := range bndl.LanguageTags() {
		currentLang := lang
		currentLangStr := currentLang.String()
		// This may be pt-BR; make it case insensitive.
		currentLangKey := strings.ToLower(strings.TrimPrefix(currentLangStr, ArtificialLangTagPrefix))
		localizer := i18n.NewLocalizer(bndl, currentLangStr)
		t.TranslateFuncs[currentLangKey] = func(ctx context.Context, translationID string, templateData any) string {
			pluralCount := valueobject.GetPluralCount(templateData)

			if templateData != nil {
				tp := reflect.TypeOf(templateData)
				if hreflect.IsInt(tp.Kind()) {
					// This was how go-i18n worked in v1,
					// and we keep it like this to avoid breaking
					// lots of sites in the wild.
					templateData = valueobject.IntCount(cast.ToInt(templateData))
				} else {
					//TODO setup with context
					if _, ok := templateData.(contenthub.Page); ok {
						// See issue 10782.
						// The i18n has its own template handling and does not know about
						// the context.Context.
						// A common pattern is to pass Page to i18n, and use .ReadingTime etc.
						// We need to improve this, but that requires some upstream changes.
						// For now, just create a wrapper.
						//templateData = page.PageWithContext{Page: p, Ctx: ctx}
					}
				}
			}

			translated, translatedLang, err := localizer.LocalizeWithTag(&i18n.LocalizeConfig{
				MessageID:    translationID,
				TemplateData: templateData,
				PluralCount:  pluralCount,
			})

			sameLang := currentLang == translatedLang

			if err == nil && sameLang {
				return translated
			}

			if err != nil && sameLang && translated != "" {
				// See #8492
				// TODO(bep) this needs to be improved/fixed upstream,
				// but currently we get an error even if the fallback to
				// "other" succeeds.
				if fmt.Sprintf("%T", err) == "i18n.pluralFormNotFoundError" {
					return translated
				}
			}

			var messageNotFoundErr *i18n.MessageNotFoundErr
			if !errors.As(err, &messageNotFoundErr) {
				t.Log.Warnf("Failed to get translated string for language %q and ID %q: %s", currentLangStr, translationID, err)
			}

			t.Log.Warnf("i18n|MISSING_TRANSLATION|%s|%s", currentLangStr, translationID)

			if enableMissingTranslationPlaceholders {
				return "[i18n] " + translationID
			}

			return translated
		}
	}
}
