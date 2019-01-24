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

package testutil

import (
	"fmt"
	"strings"
)

// CheckErrorCases checks error cases for tests
func CheckErrorCases(err error, expErrSubstr string) error {
	if err == nil && expErrSubstr != "" {
		return fmt.Errorf("got no error but expected error containing %q", expErrSubstr)
	} else if err != nil && expErrSubstr == "" {
		return fmt.Errorf("got error %q but expected no error", err.Error())
	} else if err != nil && !strings.Contains(err.Error(), expErrSubstr) {
		return fmt.Errorf("Got error %q but expected it to contain %q", err.Error(), expErrSubstr)
	}
	return nil
}
