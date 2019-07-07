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

package componentsuite

import (
	"fmt"
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/patchtmpl"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type objKey struct {
	name string
	kind string
}

// String returns the stringified objKey.
func (o objKey) String() string {
	return fmt.Sprintf("{name: %s, kind: %s}", o.name, o.kind)
}

func objKeyFromObjCheck(oc ObjectCheck) objKey {
	return objKey{
		name: oc.Name,
		kind: oc.Kind,
	}
}

func objKeyFromObj(oc *unstructured.Unstructured) objKey {
	return objKey{
		name: oc.GetName(),
		kind: oc.GetKind(),
	}
}

func runValidate(t *testing.T, comp *bundle.Component, tc *TestCase) {
	if errList := validate.Component(comp); len(errList) > 0 {
		t.Errorf("There were errors validating component:\n%v", errList.ToAggregate())
	}

	if tc.Expect.CanKubeDeserialize {
		checkObjectsDeserialize(t, comp)
	}

	checkObjectProperties(t, comp, tc)
}

func checkObjectsDeserialize(t *testing.T, comp *bundle.Component) {
	scheme := patchtmpl.DefaultPatcherScheme()
	deserializer := scheme.Codecs.UniversalDeserializer()

	for _, obj := range comp.Spec.Objects {
		key := objKeyFromObj(obj)
		objByt, err := converter.FromObject(obj).ToJSON()
		if err != nil {
			// This would be pretty unlikely
			t.Errorf("Error converting object %v to JSON: %v", key, err)
			continue
		}

		if _, err = runtime.Decode(deserializer, objByt); err != nil {
			t.Errorf("Error decode object %v: %v", key, err)
		}
	}
}

func checkObjectProperties(t *testing.T, comp *bundle.Component, tc *TestCase) {
	objCheckMap := make(map[objKey]ObjectCheck)
	for _, oc := range tc.Expect.Objects {
		objCheckMap[objKeyFromObjCheck(oc)] = oc
	}

	objMap := make(map[objKey]string)
	for _, obj := range comp.Spec.Objects {
		key := objKeyFromObj(obj)
		t.Run(fmt.Sprintf("for object %v", obj), func(t *testing.T) {
			matchFail := false
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
				t.Logf("Contents for object that didn't meet expectations %v:\n%s", key, objStr)
			}

		})
	}

	for key := range objCheckMap {
		if _, ok := objMap[key]; !ok {
			t.Errorf("Got object-keys %s, but expected to find object %v", stringMapKeys(objMap), key)
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
