package lang

import "context"

type Translator interface {
	Translate(ctx context.Context, translationID string, templateData any) string
}
