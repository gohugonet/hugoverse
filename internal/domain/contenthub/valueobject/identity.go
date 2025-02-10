package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
)

type Identity struct {
	Id uint64

	Lang    string
	LangIdx int
}

func (i Identity) ID() int {
	return int(i.Id)
}

func (i Identity) IdentifierBase() string {
	return fmt.Sprintf("%d-%s", i.Id, i.Lang)
}

func (i Identity) PageLanguage() string {
	return i.Lang
}

func (i Identity) PageLanguageIndex() int {
	return i.LangIdx
}

func (i Identity) PageIdentity() contenthub.PageIdentity {
	return i
}
