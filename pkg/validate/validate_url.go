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

package validate

import (
	"fmt"
	"net/url"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateURL validates urls in object files follwing Go's net/url parsing rules.
func validateURL(path *field.Path, u string) *field.Error {
	if u == "" {
		return field.Required(path, "url field was empty")
	}
	p, err := url.Parse(u)
	if err != nil {
		return field.Invalid(path, u, fmt.Sprintf("error parsing url: %v", err))
	}
	upath := p.Path
	// Only a sith deals in absolutes.
	if !filepath.IsAbs(upath) {
		return field.Invalid(path, upath, fmt.Sprintf("all url paths must be absolute"))
	}
	return nil
}
