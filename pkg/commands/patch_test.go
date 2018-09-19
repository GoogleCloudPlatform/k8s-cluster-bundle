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

package commands

import (
	"errors"
	"testing"

	test "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

const optionsCR = `
apiVersion: bundles/v1alpha1
kind: BundleOptions
foo: Bar
`

// Fake implementation of compFinder for unit tests.
type fakeFinder struct {
	validComp string
}

func (f *fakeFinder) ClusterComponent(name string) *bpb.ClusterComponent {
	if name == f.validComp {
		return &bpb.ClusterComponent{}
	}
	return nil
}

// Fake implementation of Patcher for unit tests.
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

// Fake implementation of OptionsReader for unit tests.
type fakeOptionsReader struct {
	validFile string
}

func (f *fakeOptionsReader) ReadOptions(filepath string) ([]byte, error) {
	if filepath == f.validFile {
		// Return an actual custom resource since we have to convert it to a custom resource map.
		return []byte(optionsCR), nil
	}
	return nil, errors.New("error reading options")
}

func TestRunPatchBundle(t *testing.T) {
	validBundleFile := "/path/to/bundle.yaml"
	validOptionsFile := "/path/to/options.yaml"
	invalidFile := "/invalid.yaml"

	testCases := []struct {
		testName string
		opts     *patchOptions
		patcher  *fakePatcher
		wantErr  bool
	}{
		{
			testName: "success case",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				output:     validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: false,
		},
		{
			testName: "bundle read error",
			opts: &patchOptions{
				bundlePath: invalidFile,
				optionsCRs: []string{validOptionsFile},
				output:     validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "options read error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile, invalidFile},
				output:     validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "bundle patching error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				output:     validBundleFile,
			},
			patcher: &fakePatcher{throwsErr: true},
			wantErr: true,
		},
		{
			testName: "bundle write error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				output:     invalidFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
	}

	or := fakeOptionsReader{validOptionsFile}
	brw := test.NewFakeReaderWriter(validBundleFile)

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Override the createPatcherFn to return a fakePatcher.
			createPatcherFn = func(b *bpb.ClusterBundle) (Patcher, error) {
				return tc.patcher, nil
			}
			err := runPatchBundle(tc.opts, brw, &or)
			if !tc.wantErr && err != nil {
				t.Errorf("runPatchBundle(opts: %+v) = error %v, want no error", tc.opts, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("runPatchBundle(opts: %+v) = no error, want error", tc.opts)
			}
		})
	}
}

func TestRunPatchApp(t *testing.T) {
	validBundleFile := "/path/to/bundle.yaml"
	validOptionsFile := "/path/to/options.yaml"
	validApp := "valid-app"
	validOutFile := "/path/to/patched.yaml"
	invalidFile := "/invalid.yaml"

	testCases := []struct {
		testName string
		opts     *patchOptions
		patcher  *fakePatcher
		wantErr  bool
	}{
		{
			testName: "success case",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  validApp,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: false,
		},
		{
			testName: "bundle read error",
			opts: &patchOptions{
				bundlePath: invalidFile,
				optionsCRs: []string{validOptionsFile},
				component:  validApp,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "options read error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile, invalidFile},
				component:  validApp,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "app not found error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  "invalid-app",
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "app patching error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  validApp,
				output:     validOutFile,
			},
			patcher: &fakePatcher{throwsErr: true},
			wantErr: true,
		},
		{
			testName: "app write error",
			opts: &patchOptions{
				bundlePath: validBundleFile,
				optionsCRs: []string{validOptionsFile},
				component:  validApp,
				output:     invalidFile,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
	}

	or := fakeOptionsReader{validOptionsFile}
	brw := test.NewFakeReaderWriter(validBundleFile)
	aw := test.NewFakeComponentWriterForPath(validOutFile)
	// Override the createFinderFn to return a fakeFinder.
	createFinderFn = func(*bpb.ClusterBundle) (compFinder, error) {
		return &fakeFinder{validComp: validApp}, nil
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Override the createPatcherFn to return a fakePatcher.
			createPatcherFn = func(*bpb.ClusterBundle) (Patcher, error) {
				return tc.patcher, nil
			}
			err := runPatchComponent(tc.opts, brw, &or, aw)
			if !tc.wantErr && err != nil {
				t.Errorf("runPatchComponent(opts: %+v) = error %v, want no error", tc.opts, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("runPatchComponent(opts: %+v) = no error, want error", tc.opts)
			}
		})
	}
}
