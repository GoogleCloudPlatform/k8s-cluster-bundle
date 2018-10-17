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
	"testing"
)

func longMaker(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += "a"
	}
	return out
}

func TestSanitizeName(t *testing.T) {
	testCases := []struct {
		desc string
		in   string
		exp  string
	}{
		{
			desc: "no change",
			in:   "foo",
			exp:  "foo",
		},
		{
			desc: "replace unknown",
			in:   "foo#o",
			exp:  "foo_o",
		},
		{
			desc: "handle normalish chars",
			in:   "foo/bar/biff.yaml",
			exp:  "foo_bar_biff.yaml",
		},
		{
			desc: "handle upper case",
			in:   "Biff",
			exp:  "biff",
		},
		{
			desc: "handle bad first/last chars",
			in:   "#buff!",
			exp:  "zbuffz",
		},
		{
			desc: "truncate long",
			in:   longMaker(260),
			exp:  longMaker(253),
		},
	}
	for _, tc := range testCases {
		out := SanitizeName(tc.in)
		if out != tc.exp {
			t.Errorf("For test %q for input %q, got %q but expected %q", tc.desc, tc.in, out, tc.exp)
		}
	}
}
