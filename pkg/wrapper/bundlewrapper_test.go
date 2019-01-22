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

// Package wrapper provides a union type for expressing various different
// bundle-types.
package wrapper

import (
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestBundleWrapper_FromRaw(t *testing.T) {
	testCases := []struct {
		desc         string
		content      string
		contentType  string
		expKind      string
		expErrSubstr string
	}{
		{
			desc: "success: bundle builder",
			content: `
apiVersion: bundle.gke.io/v1alpha1
kind: BundleBuilder`,
			contentType: "yaml",
			expKind:     "BundleBuilder",
		},
		{
			desc: "success: bundle",
			content: `
apiVersion: bundle.gke.io/v1alpha1
kind: Bundle`,
			contentType: "yaml",
			expKind:     "Bundle",
		},
		{
			desc: "success: component builder",
			content: `
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder`,
			contentType: "yaml",
			expKind:     "ComponentBuilder",
		},
		{
			desc: "success: component",
			content: `
apiVersion: bundle.gke.io/v1alpha1
kind: Component`,
			contentType: "yaml",
			expKind:     "Component",
		},
		{
			desc: "success: bundle builder json",
			content: `{
"apiVersion": "bundle.gke.io/v1alpha1",
"kind": "BundleBuilder"
}`,
			contentType: "json",
			expKind:     "BundleBuilder",
		},

		// errors
		{
			desc:         "fail: empty content",
			contentType:  "json",
			expErrSubstr: "content was empty",
		},
		{
			desc: "fail: empty content type",
			content: `{
"apiVersion": "bundle.gke.io/v1alpha1",
"kind": "BundleBuilder"
}`,
			expErrSubstr: "format was empty",
		},
		{
			desc: "fail: unrecognized kind",
			content: `{
"apiVersion": "bundle.gke.io/v1alpha1",
"kind": "BundleBlarg"
}`,
			contentType:  "json",
			expErrSubstr: "unrecognized bundle-kind",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			bw, err := FromRaw(tc.contentType, []byte(tc.content))
			testutil.CheckErrorCases(t, err, tc.expErrSubstr)
			if err != nil {
				return
			}
			if kind := bw.Kind(); tc.expKind != "" && kind != tc.expKind {
				t.Errorf("Got kind %q, but expected kind to be %q", kind, tc.expKind)
			}
		})
	}
}
