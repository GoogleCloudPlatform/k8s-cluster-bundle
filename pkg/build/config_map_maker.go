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

package build

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

type configMapMaker struct {
	cfgMap *corev1.ConfigMap
}

// Make a new ConfigMap with a metdata.name.
//
// Note that metadata.name fields have restrictions and so passed-in names will
// be sanitized.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func newConfigMapMaker(name string) *configMapMaker {
	sanitizedName := converter.SanitizeName(name)
	c := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        sanitizedName,
			Annotations: make(map[string]string),
		},
		Data: make(map[string]string),
	}
	return &configMapMaker{c}
}

// AddData adds a data-key to the config map.
func (m *configMapMaker) addData(key, value string) {
	// ConfigMaps require that each key must consist of alphanumeric characters,
	// '-', '_' or '.'.
	sanitizedKey := converter.SanitizeName(key)
	m.cfgMap.Data[sanitizedKey] = value
}

// toUnstructured converts the config map to an Unstructured type.
func (m *configMapMaker) toUnstructured() (*unstructured.Unstructured, error) {
	json, err := converter.FromObject(m.cfgMap).ToJSON()
	if err != nil {
		return nil, fmt.Errorf("error converting toUnstructured: %v", err)
	}
	return converter.FromJSON(json).ToUnstructured()
}
