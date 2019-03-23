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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

type filterType string

const (
	sel filterType = "select"
	fil filterType = "filter"
)

var example = `
components:
- spec:
    componentName: zap
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
    componentName: bog
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
    componentName: nog
    objects:
    - apiVersion: v1beta1
      kind: Pod
      metadata:
        name: nog-pod
        labels:
          component: nork
        annotations:
          foof: narf
        namespace: kube
- spec:
    componentName: zog
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
		fil         filterType
		expObjNames []string
	}{
		{
			desc: "filter-success: matches everything",
			opt:  &Options{},
			fil:  fil,
		},
		{
			desc: "filter-success: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			fil:         fil,
			expObjNames: []string{"bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
			},
			fil:         fil,
			expObjNames: []string{"zap-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
			},
			fil:         fil,
			expObjNames: []string{"zap-pod", "bog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
			},
			fil:         fil,
			expObjNames: []string{"nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: kind filter",
			opt: &Options{
				Kinds: []string{"Pod"},
			},
			fil:         fil,
			expObjNames: []string{"zog-dep"},
		},
		{
			desc: "filter-success: qualified kind filter",
			opt: &Options{
				Kinds: []string{"v1beta1/Pod"},
			},
			fil:         fil,
			expObjNames: []string{"zap-pod", "bog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: kind filter, invert",
			opt: &Options{
				Kinds:       []string{"Pod"},
				InvertMatch: true,
			},
			fil:         fil,
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod"},
		},

		// Select
		{
			desc:        "filter-success select: empty",
			opt:         &Options{},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success select: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod"},
		},
		{
			desc: "filter-success select: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
			},
			fil:         sel,
			expObjNames: []string{"bog-pod"},
		},
		{
			desc: "filter-success select: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
			},
			fil:         sel,
			expObjNames: []string{"nog-pod"},
		},
		{
			desc: "filter-success select: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod"},
		},
		{
			desc: "filter-success select: kind filter",
			opt: &Options{
				Kinds: []string{"Pod"},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod"},
		},
		{
			desc: "filter-success: qualified kind filter",
			opt: &Options{
				Kinds: []string{"v1beta1/Pod"},
			},
			fil:         sel,
			expObjNames: []string{"nog-pod"},
		},
		{
			desc: "filter-success: kind filter",
			opt: &Options{
				Kinds:       []string{"Pod"},
				InvertMatch: true,
			},
			fil:         sel,
			expObjNames: []string{"zog-dep"},
		},

		// Multiple-options filter
		{
			desc: "filter-success select: kind filter",
			opt: &Options{
				Kinds: []string{"Pod", "Deployment"}, // Pod or Deployment
				Annotations: map[string]string{
					"foof": "yar",
					"foo":  "bar",
				},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod"},
		},
	}

	data, err := converter.FromYAMLString(example).ToBundle()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			var newData []*unstructured.Unstructured
			if tc.fil == fil {
				newData = NewFilter().FilterObjects(flatten(data.Components), tc.opt)
			} else if tc.fil == sel {
				newData = NewFilter().SelectObjects(flatten(data.Components), tc.opt)
			} else {
				t.Errorf("unknown filter type %q", tc.fil)
				return
			}
			onames := getObjNames(newData)
			if !reflect.DeepEqual(onames, tc.expObjNames) {
				t.Errorf("Filter.Objects(): got %v but wanted %v", onames, tc.expObjNames)
			}
		})
	}
}

func flatten(comp []*bundle.Component) []*unstructured.Unstructured {
	var out []*unstructured.Unstructured
	for _, c := range comp {
		for _, obj := range c.Spec.Objects {
			out = append(out, obj)
		}
	}
	return out
}

func getObjNames(obj []*unstructured.Unstructured) []string {
	var names []string
	for _, o := range obj {
		names = append(names, o.GetName())
	}
	return names
}

var componentExample = `
components:
- kind: Component
  metadata:
    name: zap-pod
    labels:
      component: zork
    annotations:
      foo: bar
    namespace: kube-system
- kind: Component
  metadata:
    name: bog-pod
    labels:
      component: bork
    annotations:
      foof: yar
    namespace: kube-system
- kind: Component
  metadata:
    name: nog-pod
    labels:
      component: nork
    annotations:
      foof: narf
    namespace: kube
- kind: Component
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
		fil         filterType
		expObjNames []string
	}{
		{
			desc: "filter-success: matches everything",
			opt:  &Options{},
			fil:  fil,
		},
		{
			desc: "filter-success: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			fil:         fil,
			expObjNames: []string{"bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
			},
			fil:         fil,
			expObjNames: []string{"zap-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
			},
			fil:         fil,
			expObjNames: []string{"zap-pod", "bog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
				InvertMatch: true,
			},
			fil:         fil,
			expObjNames: []string{"nog-pod"},
		},
		{
			desc: "filter-success: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
			},
			fil:         fil,
			expObjNames: []string{"nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success: kind filter",
			opt: &Options{
				Kinds: []string{"Component"},
			},
			fil: fil,
		},

		// Select
		{
			desc:        "filter-success select: matches everything",
			opt:         &Options{},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "filter-success select: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod"},
		},
		{
			desc: "filter-success select: labels filter",
			opt: &Options{
				Labels: map[string]string{
					"component": "bork",
				},
			},
			fil:         sel,
			expObjNames: []string{"bog-pod"},
		},
		{
			desc: "filter-success select: annotations filter",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
			},
			fil:         sel,
			expObjNames: []string{"nog-pod"},
		},
		{
			desc: "filter-success select: annotations filter, invert",
			opt: &Options{
				Annotations: map[string]string{
					"foof": "narf",
				},
				InvertMatch: true,
			},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod", "zog-dep"},
		},
		{
			desc: "filter-success select: namespace filter",
			opt: &Options{
				Namespaces: []string{"kube-system"},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod"},
		},
		{
			desc: "filter-success select: kind filter",
			opt: &Options{
				Kinds: []string{"Component"},
			},
			fil:         sel,
			expObjNames: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
	}

	data, err := converter.FromYAMLString(componentExample).ToBundle()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			var newData []*bundle.Component
			if tc.fil == fil {
				newData = NewFilter().FilterComponents(data.Components, tc.opt)
			} else if tc.fil == sel {
				newData = NewFilter().SelectComponents(data.Components, tc.opt)
			} else {
				t.Errorf("unknown filter type %q", tc.fil)
				return
			}
			onames := getCompObjNames(newData)
			if !reflect.DeepEqual(onames, tc.expObjNames) {
				t.Errorf("FilterComponents(): got %v but wanted %v", onames, tc.expObjNames)
			}
		})
	}
}

func getCompObjNames(comp []*bundle.Component) []string {
	var names []string
	for _, c := range comp {
		names = append(names, c.ObjectMeta.Name)
	}
	return names
}

func TestPartitionObjects(t *testing.T) {
	testcases := []struct {
		desc        string
		opt         *Options
		expMatch    []string
		expNotMatch []string
	}{
		{
			desc:     "match everything",
			opt:      &Options{},
			expMatch: []string{"zap-pod", "bog-pod", "nog-pod", "zog-dep"},
		},
		{
			desc: "match single pod: name filter",
			opt: &Options{
				Names: []string{"zap-pod"},
			},
			expMatch:    []string{"zap-pod"},
			expNotMatch: []string{"bog-pod", "nog-pod", "zog-dep"},
		},
	}

	data, err := converter.FromYAMLString(example).ToBundle()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			match, notMatch := NewFilter().PartitionObjects(flatten(data.Components), tc.opt)
			matchNames := getObjNames(match)
			notMatchNames := getObjNames(notMatch)
			if !reflect.DeepEqual(matchNames, tc.expMatch) {
				t.Errorf("Filter.PartitionObjects(): got match group %v but wanted %v", matchNames, tc.expMatch)
			}
			if !reflect.DeepEqual(notMatchNames, tc.expNotMatch) {
				t.Errorf("Filter.PartitionObjects(): got not-match group %v but wanted %v", notMatchNames, tc.expNotMatch)
			}
		})
	}
}
