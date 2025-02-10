package application

import (
	"github.com/mdfriday/hugoverse/pkg/testkit"
)

func NewDemo() (string, error) {
	tmpDir, _, err := testkit.MkBookSite()
	//defer clean()

	if err != nil {
		return "", err
	}

	return tmpDir, nil
}
