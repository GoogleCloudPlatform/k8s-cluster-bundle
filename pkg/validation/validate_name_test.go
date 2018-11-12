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
	"strings"
	"testing"
)

func longNameMaker(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += "a"
	}
	return out
}

func TestValidateName(t *testing.T) {
	testCases := []struct {
		desc      string
		name      string
		errSubStr string
	}{
		{
			desc: "success",
			name: "foo",
		}, {
			desc:      "bad first char",
			name:      "*oo",
			errSubStr: "first character",
		}, {
			desc:      "bad last char",
			name:      "foo*",
			errSubStr: "last character",
		}, {
			desc:      "bad middle char",
			name:      "f*oo",
			errSubStr: "allowed characters",
		}, {
			desc:      "empty name",
			errSubStr: "was empty",
		}, {
			desc:      "too long name",
			name:      longNameMaker(254),
			errSubStr: "was longer",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := ValidateName(tc.name)
			if err == nil {
				if tc.errSubStr != "" && !strings.Contains(err.Error(), tc.errSubStr) {
					t.Errorf("got nil error, but expected one with %q", tc.errSubStr)
					return
				}
				return // no error, not a problem
			} else {
				if tc.errSubStr == "" {
					t.Errorf("got error %q, but didn't expect one", err.Error())
					return
				} else if !strings.Contains(err.Error(), tc.errSubStr) {
					t.Errorf("got error %q, but expected it to contain %q", err.Error(), tc.errSubStr)
					return
				}
				return // error matches
			}
		})
	}
}
