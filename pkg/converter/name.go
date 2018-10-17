// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"regexp"
	"strings"
)

var firstCharRegexp = regexp.MustCompile(`^[^a-z0-9]`)
var nameRegexp = regexp.MustCompile(`[^a-z0-9_.-]`)
var lastCharRegexp = regexp.MustCompile(`[^a-z0-9]$`)

// SanitizeName sanitizes a metadata.name field, replacing unsafe characters
// with _ and truncating if it's longer than 253 characters.
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	name = nameRegexp.ReplaceAllString(name, "_")
	name = firstCharRegexp.ReplaceAllString(name, "z")
	name = lastCharRegexp.ReplaceAllString(name, "z")
	if len(name) >= 254 {
		name = name[0:253]
	}
	return name
}
