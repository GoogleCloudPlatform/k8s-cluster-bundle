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
	"strings"
	"testing"
)

func TestDecodeBundle(t *testing.T) {
	testCases := []struct {
		desc               string
		data               string
		allowUnknownFields bool
		format             ContentType
		expErrSubstr       string
	}{
		{
			desc: "Success decode: yaml",
			data: `
kind: Bundle
apiVersion: bundle.gke.io/v1alpha1
metadata:
  name: much-bundle
`,
			format: YAML,
		},
		{
			desc: "Success decode: json",
			data: `
{
  "kind": "bundle",
  "apiVersion": "bundle.gke.io/v1alpha1",
  "metadata": {
    "name": "much-bundle"
  }
}`,
			format: JSON,
		},
		{
			desc: "Success decode: json as yaml",
			data: `
{
  "kind": "bundle",
  "apiVersion": "bundle.gke.io/v1alpha1",
  "metadata": {
    "name": "much-bundle"
  }
}`,
			format: YAML,
		},
		{
			desc: "Error decode: yaml as json",
			data: `
kind: Bundle
apiVersion: bundle.gke.io/v1alpha1
metadata:
  name: much-bundle
`,
			format:       JSON,
			expErrSubstr: "invalid character",
		},
		{
			desc: "Error decode: bad format",
			data: `
kind: Bundle
apiVersion: bundle.gke.io/v1alpha1
metadata:
  name: much-bundle
`,
			format:       ContentType("zop"),
			expErrSubstr: "unknown content type",
		},
		{
			desc: "Error decode: unknown fields",
			data: `
kind: Bundle
apiVersion: bundle.gke.io/v1alpha1
metadata:
  name: much-bundle
zorp: zip
`,
			format:       YAML,
			expErrSubstr: "unknown field \"zorp\"",
		},
		{
			desc: "Success decode: unknown fields + allowUnknownFields",
			data: `
kind: Bundle
apiVersion: bundle.gke.io/v1alpha1
metadata:
  name: much-bundle
zorp: zip
`,
			format:             YAML,
			allowUnknownFields: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &Decoder{
				data:               []byte(tc.data),
				allowUnknownFields: tc.allowUnknownFields,
				format:             tc.format,
			}

			b, err := m.ToBundle()
			if err == nil && tc.expErrSubstr == "" {
				// success
			} else if err == nil && tc.expErrSubstr != "" {
				t.Fatalf("Got nil err, but expected one containing %s", tc.expErrSubstr)
			} else if err != nil && tc.expErrSubstr == "" {
				t.Fatalf("Got error %s, but did not expect one", err.Error())
			} else if err != nil && tc.expErrSubstr != "" && !strings.Contains(err.Error(), tc.expErrSubstr) {
				t.Fatalf("Got error %q, but expected it to contain substring %q", err.Error(), tc.expErrSubstr)
			} else if err != nil {
				// Success, but we can't continue
				return
			}
			if b == nil {
				t.Fatalf("Got nil bundle, but expected it to be non-nil")
			}

			if b.ObjectMeta.Name != "much-bundle" {
				t.Errorf("got name %q, expected it to contain %q", b.ObjectMeta.Name, "much-bundle")
			}
		})
	}
}
