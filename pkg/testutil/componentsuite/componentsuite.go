// Copyright 2019 Google LLC
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

// Package componentsuite provides a test-suite helper for running component tests.
package componentsuite

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/build"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/multi"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
)

// Run runs a component-tester test-suite. testSuiteFile specifies the path to
// the a component-test-suite YAML file.
//
// By default, go runs tests with the cwd being the current directory.
func Run(t *testing.T, testSuiteFile string) {
	t.Logf("Running tests for suite %q", testSuiteFile)

	ctx := context.Background()

	data, err := ioutil.ReadFile(testSuiteFile)
	if err != nil {
		t.Fatalf("while reading test suite file %q, %v", testSuiteFile, err)
	}

	ts := &ComponentTestSuite{}
	if err := converter.FromFileName(testSuiteFile, data).ToObject(ts); err != nil {
		t.Fatalf("while parsing test-suite file %q, %v", testSuiteFile, err)
	}

	os.Chdir(filepath.Join(filepath.Dir(testSuiteFile), ts.RootDirectory))

	// Inlining expects an absolute parent path, so make the path absolute now.
	cfile := ts.ComponentFile
	if !filepath.IsAbs(cfile) {
		cfile, err = filepath.Abs(cfile)
		if err != nil {
			t.Fatalf("While making an absolute path out of %s: %v", cfile, err)
		}
	}

	cdata, err := ioutil.ReadFile(cfile)
	if err != nil {
		t.Fatalf("While reading component file %q, %v", cfile, err)
	}

	bw, err := wrapper.FromRaw(string(converter.YAML), cdata)
	if err != nil {
		t.Fatalf("While converting component file %q, %v", cfile, err)
	}

	if bw.Kind() != "Component" && bw.Kind() != "ComponentBuilder" {
		t.Fatalf("Got kind %q, but component kind must be \"Component\" or \"ComponentBuilder\"", bw.Kind())
	}

	var comp *bundle.Component
	if bw.Kind() == "ComponentBuilder" {
		inliner := build.NewLocalInliner(ts.RootDirectory)
		comp, err = inliner.ComponentFiles(ctx, bw.ComponentBuilder(), cfile)
		if err != nil {
			t.Fatalf("Inlining component files: %v", err)
		}
	} else {
		comp = bw.Component()
	}

	for _, tc := range ts.TestCases {
		tci := tc // necessary.
		t.Run(tc.Description, func(t *testing.T) {
			t.Parallel()
			runTest(t, comp.DeepCopy(), tci)
		})
	}
}

func runTest(t *testing.T, comp *bundle.Component, tc *TestCase) {
	comp = runBuild(t, comp, tc)
	if comp == nil {
		return
	}

	comp = runApply(t, comp, tc)
	if comp == nil {
		return
	}

	runValidate(t, comp, tc)
}

func runBuild(t *testing.T, comp *bundle.Component, tc *TestCase) *bundle.Component {
	buildFilter := &filter.Options{}
	comp, err := build.ComponentPatchTemplates(comp, buildFilter, tc.Build.Options)
	cerr := testutil.CheckErrorCases(err, tc.Expect.BuildErrSubstr)
	if cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		// Since there's an error, it's a terminal condition. We can't proceed.
		return nil
	}
	return comp
}

func runApply(t *testing.T, comp *bundle.Component, tc *TestCase) *bundle.Component {
	applier := multi.NewDefaultApplier()

	comp, err := applier.ApplyOptions(comp, tc.Apply.Options)
	cerr := testutil.CheckErrorCases(err, tc.Expect.ApplyErrSubstr)
	if cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		// Since there's an error, it's a terminal condition. We've already checked
		// that the err is as expected in CheckErrorCases.
		return nil
	}
	return comp
}
