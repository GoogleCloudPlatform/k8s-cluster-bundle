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

	"k8s.io/apimachinery/pkg/util/validation/field"
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
			errSubStr: "must consist of",
		}, {
			desc:      "bad last char",
			name:      "foo*",
			errSubStr: "must consist of",
		}, {
			desc:      "bad middle char",
			name:      "f*oo",
			errSubStr: "must consist of",
		}, {
			desc:      "empty name",
			errSubStr: "must consist of",
		}, {
			desc:      "too long name",
			name:      longNameMaker(254),
			errSubStr: "must be no more than 63 characters",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateName(field.NewPath("testpath"), tc.name).ToAggregate()
			if err == nil && tc.errSubStr != "" {
				t.Fatalf("got nil error, but expected one with %q", tc.errSubStr)
			}
			if err != nil && tc.errSubStr == "" {
				t.Fatalf("got error %q, but didn't expect one", err.Error())
			}
			if err != nil && !strings.Contains(err.Error(), tc.errSubStr) {
				t.Fatalf("got error %q, but expected it to contain %q", err.Error(), tc.errSubStr)
			}
		})
	}
}
