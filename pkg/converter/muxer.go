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
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// FromYAMLString creates a Muxer instance from a YAML string.
func FromYAMLString(s string) *Muxer {
	return FromYAML([]byte(s))
}

// FromYAML creates a Muxer instance from a YAML byte array.
func FromYAML(b []byte) *Muxer {
	return &Muxer{
		data:   b,
		format: YAML,
	}
}

// FromJSONString creates a Muxer instance from a JSON string.
func FromJSONString(s string) *Muxer {
	return FromJSON([]byte(s))
}

// FromJSON creates a Muxer instance from a JSON byte array.
func FromJSON(b []byte) *Muxer {
	return &Muxer{
		data:   b,
		format: YAML,
	}
}

// FromFileName creates a Muxer instance by guessing the type based on the file
// extension.
func FromFileName(fname string, contents []byte) *Muxer {
	ext := filepath.Ext(fname)
	switch ext {
	case ".yaml":
		return FromYAML(contents)
	case ".json":
		return FromJSON(contents)
	default:
		// This will be an error during conversion
		return &Muxer{format: UnknownContent}
	}
}

// FromContentType takes an explicit content type to use for creating a
// conversion Muxer.
func FromContentType(ctype string, contents []byte) *Muxer {
	ctype = strings.ToLower(ctype)
	return &Muxer{format: ContentType(ctype), data: contents}
}

// Muxer converts from an object's serialized format to an actual instance of
// the object.
type Muxer struct {
	data   []byte
	format ContentType
}

func (m *Muxer) mux(f interface{}) error {
	switch m.format {
	case YAML:
		return yaml.Unmarshal(m.data, f)
	case JSON:
		return json.Unmarshal(m.data, f)
	default:
		return fmt.Errorf("Unknown content type: %q", m.format)
	}
}

// ToBundle converts input data to the Bundle type.
func (m *Muxer) ToBundle() (*bundle.Bundle, error) {
	d := &bundle.Bundle{}
	if err := m.mux(d); err != nil {
		return nil, err
	}
	return d, nil
}

// ToComponentPackage converts input data to the ComponentPackage type.
func (m *Muxer) ToComponentPackage() (*bundle.ComponentPackage, error) {
	d := &bundle.ComponentPackage{}
	if err := m.mux(d); err != nil {
		return nil, err
	}
	return d, nil
}

// ToComponentSet converts input data to the ComponentSet type.
func (m *Muxer) ToComponentSet() (*bundle.ComponentSet, error) {
	d := &bundle.ComponentSet{}
	if err := m.mux(d); err != nil {
		return nil, err
	}
	return d, nil
}

// ToUnstructured converts input data to the Unstructured type.
func (m *Muxer) ToUnstructured() (*unstructured.Unstructured, error) {
	d := &unstructured.Unstructured{}
	if err := m.mux(d); err != nil {
		return nil, err
	}
	return d, nil
}

// ToObject converts input data to the correct structured type.
func (m *Muxer) ToObject() (runtime.Object, error) {
	// TODO: I think we can just use a Scheme for this
	u, err := m.ToUnstructured()
	if err != nil {
		return nil, err
	}

	kind := u.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case "Bundle":
		return m.ToBundle()
	case "ComponentPackage":
		return m.ToComponentPackage()
	case "ComponentSet":
		return m.ToComponentSet()
	default:
		return nil, fmt.Errorf("unknown Kind: %q", kind)
	}
}
