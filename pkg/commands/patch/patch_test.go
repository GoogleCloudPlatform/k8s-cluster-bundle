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

package patch

import (
	"context"
	"errors"
	"strings"
	"testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

const optionsCR = `
apiVersion: bundles/v1alpha1
kind: BundleOptions
foo: Bar
`

// Fake implementation of componentFinder for unit tests.
type fakeFinder struct {
	validComp string
}

func (f *fakeFinder) ClusterComponent(name string) *bpb.ClusterComponent {
	if name == f.validComp {
		return &bpb.ClusterComponent{}
	}
	return nil
}

// Fake implementation of patcher for unit tests.
type fakePatcher struct {
	throwsErr bool
}

func (f *fakePatcher) PatchBundle([]map[string]interface{}) (*bpb.ClusterBundle, error) {
	if f.throwsErr {
		return nil, errors.New("error patching bundle")
	}
	return &bpb.ClusterBundle{}, nil
}

func (f *fakePatcher) PatchComponent(*bpb.ClusterComponent, []map[string]interface{}) (*bpb.ClusterComponent, error) {
	if f.throwsErr {
		return nil, errors.New("error patching component")
	}
	return &bpb.ClusterComponent{}, nil
}

func TestRunPatchBundle(t *testing.T) {
	validBundleFile := "/path/to/bundle.yaml"
	validOptionsFile := "/path/to/options.yaml"
	invalidFile := "/invalid.yaml"

	testCases := []struct {
		testName string
		opts     *options
		patcher  *fakePatcher
		wantErr  bool
	}{
		{
			testName: "success case",
			opts: &options{
				bundlePath: "in" + validBundleFile,
				optionsCRs: []string{validOptionsFile},
				output:     "out" + validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: false,
		},
		{
			testName: "bundle read error",
			opts: &options{
				bundlePath: "in" + invalidFile,
				optionsCRs: []string{validOptionsFile},
				output:     "out" + validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "options read error",
			opts: &options{
				bundlePath: "in" + validBundleFile,
				optionsCRs: []string{validOptionsFile, invalidFile},
				output:     "out" + validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "bundle patching error",
			opts: &options{
				bundlePath: "in" + validBundleFile,
				optionsCRs: []string{validOptionsFile},
				output:     "out" + validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: true},
			wantErr: true,
		},
		{
			testName: "bundle write error",
			opts: &options{
				bundlePath: "out" + validBundleFile,
				optionsCRs: []string{validOptionsFile},
				output:     "out" + invalidFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			var pairs []*testutil.FilePair
			if strings.Contains(tc.opts.bundlePath, validBundleFile) {
				pairs = append(pairs, &testutil.FilePair{tc.opts.bundlePath, testutil.FakeBundle})
			}
			if strings.Contains(tc.opts.output, validBundleFile) {
				pairs = append(pairs, &testutil.FilePair{tc.opts.output, testutil.FakeBundle})
			}
			for _, o := range tc.opts.optionsCRs {
				if strings.Contains(o, validOptionsFile) {
					pairs = append(pairs, &testutil.FilePair{o, testutil.FakeBundle})
				}
			}

			rw := testutil.NewFakeReaderWriterFromPairs(pairs...)
			// Override the createPatcherFn to return a fakePatcher.
			createPatcherFn = func(b *bpb.ClusterBundle) (patcher, error) {
				return tc.patcher, nil
			}
			err := runPatchBundle(ctx, tc.opts, rw)
			if !tc.wantErr && err != nil {
				t.Errorf("runPatchBundle(opts: %+v) = error %v, want no error", tc.opts, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("runPatchBundle(opts: %+v) = no error, want error", tc.opts)
			}
		})
	}
}

func TestRunPatchComponent(t *testing.T) {
	validBundleFile := "/path/to/bundle.yaml"
	validOptionsFile := "/path/to/options.yaml"
	validComponent := "valid-component"
	validOutFile := "/path/to/patched.yaml"
	invalidFile := "/invalid.yaml"

	testCases := []struct {
		testName string
		opts     *options
		patcher  *fakePatcher
		wantErr  bool
	}{
		{
			testName: "success case",
			opts: &options{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: false,
		},
		{
			testName: "bundle read error",
			opts: &options{
				bundlePath: invalidFile,
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "options read error",
			opts: &options{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile, invalidFile},
				component:  validComponent,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "component not found error",
			opts: &options{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  "invalid-component",
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "component patching error",
			opts: &options{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: true},
			wantErr: true,
		},
		{
			testName: "component write error",
			opts: &options{
				bundlePath: "in" + validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
				output:     "out" + invalidFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
	}

	// Override the createFinderFn to return a fakeFinder.
	createFinderFn = func(*bpb.ClusterBundle) (componentFinder, error) {
		return &fakeFinder{validComp: validComponent}, nil
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			var pairs []*testutil.FilePair
			if strings.Contains(tc.opts.bundlePath, validBundleFile) {
				pairs = append(pairs, &testutil.FilePair{tc.opts.bundlePath, testutil.FakeBundle})
			}
			if strings.Contains(tc.opts.output, validOutFile) {
				pairs = append(pairs, &testutil.FilePair{tc.opts.output, testutil.FakeBundle})
			}
			for _, o := range tc.opts.optionsCRs {
				if strings.Contains(o, validOptionsFile) {
					pairs = append(pairs, &testutil.FilePair{o, testutil.FakeBundle})
				}
			}

			rw := testutil.NewFakeReaderWriterFromPairs(pairs...)

			// Override the createPatcherFn to return a fakePatcher.
			createPatcherFn = func(*bpb.ClusterBundle) (patcher, error) {
				return tc.patcher, nil
			}
			err := runPatchComponent(ctx, tc.opts, rw)
			if !tc.wantErr && err != nil {
				t.Errorf("runPatchComponent(opts: %+v) = error %v, want no error", tc.opts, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("runPatchComponent(opts: %+v) = no error, want error", tc.opts)
			}
		})
	}
}
