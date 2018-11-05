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

package filter

import (
	"reflect"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

var example = `
components:
- spec:
    canonicalName: zap
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: zap-pod
        labels:
          component: zork
        annotations:
          foo: bar
        namespace: kube-system
- spec:
    canonicalName: bog
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: bog-pod
        labels:
          component: bork
        annotations:
          foof: yar
        namespace: kube-system
- spec:
    canonicalName: nog
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: nog-pod
        labels:
          component: nork
        annotations:
          foof: narf
        namespace: kube
- spec:
    canonicalName: zog
    objects:
    - apiVersion: v1
      kind: Deployment
      metadata:
        name: zog-dep
        labels:
          component: zork
        annotations:
          zoof: zarf
        namespace: zube`

func TestFilterObjects(t *testing.T) {
	testcases := []struct {
		desc        string
		opt         *Options
		expObjNames []string
	}{
		{
			desc:        "fiter-success: no change",
			opt:         &Options{},
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			expObjNames: []string{"bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
			},
			expObjNames: []string{"zap-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
			},
			expObjNames: []string{"zap-pod", "bog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
			},
			expObjNames: []string{"nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: kind filter",
			opt: &Options{
				Kinds: []string{"Pod"},
			},
			expObjNames: []string{"zog-dep"},
		},

		// KeepOnly
		{
			desc: "fiter-success keeponly: empty",
			opt: &Options{
				KeepOnly: true,
			},
		},
		{
			desc: "fiter-success keeponly: name filter",
			opt: &Options{
				Names:    []string{"zap-pod"},
				KeepOnly: true,
			},
			expObjNames: []string{"zap-pod"},
		},
		{
			desc: "fiter-success keeponly: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
				KeepOnly: true,
			},
			expObjNames: []string{"bog-pod"},
		},
		{
			desc: "fiter-success keeponly: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
				KeepOnly: true,
			},
			expObjNames: []string{"nog-pod"},
		},
		{
			desc: "fiter-success keeponly: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
				KeepOnly:   true,
			},
			expObjNames: []string{"zap-pod", "bog-pod"},
		},
		{
			desc: "fiter-success keeponly: kind filter",
			opt: &Options{
				Kinds:    []string{"Pod"},
				KeepOnly: true,
			},
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod"},
		},
	}

	data, err := converter.FromYAMLString(example).ToComponentData()
	if err != nil {
		t.Fatalf("error converting component data: %v", err)
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			f := &Filterer{data.Components}
			newData := f.FilterObjects(tc.opt)
			onames := getObjNames(newData)
			if !reflect.DeepEqual(onames, tc.expObjNames) {
				t.Errorf("FilterObjects(): got %v but wanted %v", onames, tc.expObjNames)
			}
		})
	}
}

func getObjNames(comp []*bundle.ComponentPackage) []string {
	var names []string
	for _, c := range comp {
		for _, o := range c.Spec.Objects {
			names = append(names, o.GetName())
		}
	}
	return names
}

var componentExample = `
components:
- kind: ComponentPackage
  metadata:
    name: zap-pod
    labels:
      component: zork
    annotations:
      foo: bar
    namespace: kube-system
- kind: ComponentPackage
  metadata:
    name: bog-pod
    labels:
      component: bork
    annotations:
      foof: yar
    namespace: kube-system
- kind: ComponentPackage
  metadata:
    name: nog-pod
    labels:
      component: nork
    annotations:
      foof: narf
    namespace: kube
- kind: ComponentPackage
  metadata:
    name: zog-dep
    labels:
      component: zork
    annotations:
      zoof: zarf
    namespace: zube`

func TestFilterComponents(t *testing.T) {
	testcases := []struct {
		desc        string
		opt         *Options
		expObjNames []string
	}{
		{
			desc:        "fiter-success: no change",
			opt:         &Options{},
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			expObjNames: []string{"bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
			},
			expObjNames: []string{"zap-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
			},
			expObjNames: []string{"zap-pod", "bog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
			},
			expObjNames: []string{"nog-pod", "zog-dep"},
		},
		{
			desc: "fiter-success: kind filter",
			opt: &Options{
				Kinds: []string{"ComponentPackage"},
			},
		},

		// KeepOnly
		{
			desc: "fiter-success keeponly: empty",
			opt: &Options{
				KeepOnly: true,
			},
		},
		{
			desc: "fiter-success keeponly: name filter",
			opt: &Options{
				Names:    []string{"zap-pod"},
				KeepOnly: true,
			},
			expObjNames: []string{"zap-pod"},
		},
		{
			desc: "fiter-success keeponly: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
				KeepOnly: true,
			},
			expObjNames: []string{"bog-pod"},
		},
		{
			desc: "fiter-success keeponly: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
				KeepOnly: true,
			},
			expObjNames: []string{"nog-pod"},
		},
		{
			desc: "fiter-success keeponly: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
				KeepOnly:   true,
			},
			expObjNames: []string{"zap-pod", "bog-pod"},
		},
		{
			desc: "fiter-success keeponly: kind filter",
			opt: &Options{
				Kinds:    []string{"ComponentPackage"},
				KeepOnly: true,
			},
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
	}

	data, err := converter.FromYAMLString(componentExample).ToComponentData()
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			f := &Filterer{data.Components}
			newData := f.FilterComponents(tc.opt)
			onames := getCompObjNames(newData)
			if !reflect.DeepEqual(onames, tc.expObjNames) {
				t.Errorf("FilterObjects(): got %v but wanted %v", onames, tc.expObjNames)
			}
		})
	}
}

func getCompObjNames(comp []*bundle.ComponentPackage) []string {
	var names []string
	for _, c := range comp {
		names = append(names, c.ObjectMeta.Name)
	}
	return names
}
