// Copyright 2021 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package htime

import (
	"time"

	clock "github.com/bep/clocks"
)

var (
	Clock = clock.System()
)

// Now returns time.Now() or time value based on the `clock` flag.
// Use this function to fake time inside hugo.
func Now() time.Time {
	return Clock.Now()
}

// AsTimeProvider is implemented by go-toml's LocalDate and LocalDateTime.
type AsTimeProvider interface {
	AsTime(zone *time.Location) time.Time
}
