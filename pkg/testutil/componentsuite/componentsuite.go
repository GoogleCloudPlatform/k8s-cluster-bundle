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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/build"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/patchtmpl"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
)

// Run runs a component-tester test-suite. testSuiteFile specifies the path to
// the current file.
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

	cdata, err := ioutil.ReadFile(ts.ComponentFile)
	if err != nil {
		t.Fatalf("while reading component file %q, %v", ts.ComponentFile, err)
	}

	bw, err := wrapper.FromRaw(string(converter.YAML), cdata)
	if err != nil {
		t.Fatalf("while converting component file %q, %v", ts.ComponentFile, err)
	}

	if bw.Kind() != "Component" && bw.Kind() != "ComponentBuilder" {
		t.Fatalf("Got kind %q, but component kind must be \"Component\" or \"ComponentBuilder\"", bw.Kind())
	}

	var comp *bundle.Component
	if bw.Kind() == "ComponentBuilder" {
		inliner := build.NewLocalInliner(ts.RootDirectory)
		comp, err = inliner.ComponentFiles(ctx, bw.ComponentBuilder())
		if err != nil {
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
	comp, err := build.BuildComponentPatchTemplates(comp, buildFilter, tc.Build.Options)
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
	applyOpts := &filter.Options{}
	applyOpts.Annotations = tc.Apply.Filters

	applier := patchtmpl.NewApplier(
		patchtmpl.DefaultPatcherScheme(),
		applyOpts,
		true /* includeTemplates, for debugging */,
		"error")

	comp, err := applier.ApplyOptions(comp, tc.Apply.Options)
	cerr := testutil.CheckErrorCases(err, tc.Expect.ApplyErrSubstr)
	if cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		// Since there's an error, it's a terminal condition
		return nil
	}
	return comp
}

type objKey struct {
	name string
	kind string
}

func (ok objKey) String() string {
	return fmt.Sprintf("{name: %q, kind: %q}", ok.name, ok.kind)
}

func runValidate(t *testing.T, comp *bundle.Component, tc *TestCase) {
	if errList := validate.Component(comp); len(errList) > 0 {
		t.Errorf("There were errors validating component:\n%v", errList.ToAggregate())
	}

	objCheckMap := make(map[objKey]ObjectCheck)
	for _, oc := range tc.Expect.Objects {
		objCheckMap[objKey{
			name: oc.Name,
			kind: oc.Kind,
		}] = oc
	}

	objMap := make(map[objKey]string)
	for _, obj := range comp.Spec.Objects {
		matchFail := false

		key := objKey{name: obj.GetName(), kind: obj.GetKind()}
		objStr, err := converter.FromObject(obj).ToYAMLString()
		if err != nil {
			// This is a very unlikely error.
			t.Fatal(err)
		}
		objMap[key] = objStr

		check := objCheckMap[key]
		for _, expStr := range check.FindSubstrs {
			if !strings.Contains(objStr, expStr) {
				t.Errorf("Did not find %q in object %v, but expected to", expStr, key)
				matchFail = true
			}
		}

		for _, noExpStr := range check.NotFindSubstrs {
			if strings.Contains(objStr, noExpStr) {
				t.Errorf("Found %q in object %v, but did not expect to", noExpStr, key)
				matchFail = true
			}
		}

		if matchFail {
			t.Errorf("Contents for object that didn't meet expectations %v:\n%s", key, objStr)
		}
	}

	for key := range objCheckMap {
		if _, ok := objMap[key]; !ok {
			t.Errorf("Got object-keys %q, but expected to find object %v", stringMapKeys(objMap), key)
		}
	}
}

func stringMapKeys(m map[objKey]string) []string {
	var out []string
	for k := range m {
		out = append(out, k.String())
	}
	return out
}
