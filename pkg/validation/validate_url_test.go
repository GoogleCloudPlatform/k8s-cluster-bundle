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

func TestValidateURL(t *testing.T) {
	testCases := []struct {
		desc         string
		in           string
		expErrSubstr string
	}{
		{
			desc: "success: normal file url",
			in:   "file:///foo/bar",
		},
		{
			desc: "success: normal file url with localhost",
			in:   "file://localhost/foo/bar",
		},
		{
			desc:         "fail: empty URL",
			in:           "",
			expErrSubstr: "url field was empty",
		},
		{
			desc:         "fail: bad parsing",
			in:           "file@://foo",
			expErrSubstr: "error parsing url",
		},
		{
			desc: "success: no scheme",
			in:   "foo/bar/biff",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateURL(field.NewPath("testpath"), tc.in)
			if err != nil {
				if tc.expErrSubstr == "" {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tc.expErrSubstr) {
					t.Fatalf("error %v did not contain expected substring %q", err, tc.expErrSubstr)
				}
				return // expected error
			}
			if tc.expErrSubstr != "" {
				t.Fatalf("got nil error, but expected error with substring %q", tc.expErrSubstr)
			}
		})
	}
}
