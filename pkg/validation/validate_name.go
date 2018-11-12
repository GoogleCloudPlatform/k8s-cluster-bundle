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

package validation

import (
	"fmt"
	"regexp"
)

var firstCharRegexp = regexp.MustCompile(`^[a-z0-9]`)
var nameRegexp = regexp.MustCompile(`^[a-z0-9_.-]+$`)
var lastCharRegexp = regexp.MustCompile(`[a-z0-9]$`)

// ValidateName validates names to ensure they follow the Kubernetes
// conventions for how names are constructed. For more about names,
// see: k8s.io/docs/concepts/overview/working-with-objects/names/
func ValidateName(n string) error {
	if n == "" {
		return fmt.Errorf("name field was empty")
	}
	if len(n) >= 254 {
		return fmt.Errorf("name %q was longer than 253 characters", n)
	}
	if !firstCharRegexp.MatchString(n) {
		return fmt.Errorf("name %q did not have a first character with the pattern %q", n, firstCharRegexp.String())
	}
	if !lastCharRegexp.MatchString(n) {
		return fmt.Errorf("name %q did not must a last character with the pattern %q", n, lastCharRegexp.String())
	}
	if !nameRegexp.MatchString(n) {
		return fmt.Errorf("name %q did not match allowed characters %q", n, nameRegexp.String())
	}

	return nil
}
