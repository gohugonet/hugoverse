package valueobject

import "errors"

const errInvalidInfo = `An unknown revision most likely means that someone has deleted the remote ref (e.g. with a force push to GitHub).
To resolve this, you need to manually edit your go.mod file and replace the version for the module in question with a valid ref.

The easiest is to just enter a valid branch name there, e.g. master, which would be what you put in place of 'v0.5.1' in the example below.

require github.com/gohugoio/hugo-mod-jslibs/instantpage v0.5.1

If you then run 'hugo mod graph' it should resolve itself to the most recent version (or commit if no semver versions are available).`

var ErrNotExist = errors.New("module does not exist")
