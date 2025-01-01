// Copyright © 2022 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package libsasserrors holds the error types used by the libsass package.
package libsasserrors

import (
	"encoding/json"
	"fmt"
)

// JsonToError converts a JSON string to an error.
func JsonToError(jsonstr string) (e Error) {
	_ = json.Unmarshal([]byte(jsonstr), &e)
	return
}

// Error is a libsass error.
type Error struct {
	Status  int    `json:"status"`
	Column  int    `json:"column"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("file %q, line %d, col %d: %s ", e.File, e.Line, e.Column, e.Message)
}
