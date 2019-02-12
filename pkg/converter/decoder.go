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
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// FromYAMLString creates a Decoder instance from a YAML string.
func FromYAMLString(s string) *Decoder {
	return FromYAML([]byte(s))
}

// FromYAML creates a Decoder instance from a YAML byte array.
func FromYAML(b []byte) *Decoder {
	return &Decoder{
		data:   b,
		format: YAML,
	}
}

// FromJSONString creates a Decoder instance from a JSON string.
func FromJSONString(s string) *Decoder {
	return FromJSON([]byte(s))
}

// FromJSON creates a Decoder instance from a JSON byte array.
func FromJSON(b []byte) *Decoder {
	return &Decoder{
		data:   b,
		format: YAML,
	}
}

// FromFileName creates a Decoder instance by guessing the type based on the file
// extension.
func FromFileName(fname string, contents []byte) *Decoder {
	ext := filepath.Ext(fname)
	var d *Decoder
	switch ext {
	case ".yaml", ".yml":
		d = FromYAML(contents)
	case ".json":
		d = FromJSON(contents)
	default:
		// This will be an error during conversion
		d = &Decoder{format: UnknownContent}
	}
	d.fname = fname
	return d
}

// FromContentType takes an explicit content type to use for creating a
// conversion Decoder.
func FromContentType(ctype string, contents []byte) *Decoder {
	ctype = strings.ToLower(ctype)
	return &Decoder{format: ContentType(ctype), data: contents}
}

// Decoder converts from an object's serialized format to an actual instance of
// the object. By default, the Decoder does not allow unknown fields.
type Decoder struct {
	data               []byte
	format             ContentType
	allowUnknownFields bool

	// If the content came from FromFileName, store the filename and report in
	// errors for debugging.
	fname string
}

// AllowUnknownFields indicates whether to allow unknown fields during decoding.
func (m *Decoder) AllowUnknownFields(allow bool) *Decoder {
	return &Decoder{
		data:               m.data,
		format:             m.format,
		allowUnknownFields: allow,
	}
}

func (m *Decoder) decode(f interface{}) error {
	var err error
	switch m.format {
	case YAML:
		var mod yaml.JSONOpt = func(d *json.Decoder) *json.Decoder {
			if !m.allowUnknownFields {
				d.DisallowUnknownFields()
			}
			return d
		}
		err = yaml.Unmarshal(m.data, f, mod)
	case JSON:
		jsonDecoder := json.NewDecoder(bytes.NewReader(m.data))
		if !m.allowUnknownFields {
			jsonDecoder.DisallowUnknownFields()
		}
		err = jsonDecoder.Decode(f)
	default:
		err = fmt.Errorf("unknown content type: %q", m.format)
	}
	if err != nil && m.fname != "" {
		return fmt.Errorf("while decoding contents from file %v, %v", m.fname, err)
	}
	return err
}

// ToBundle converts input data to the Bundle type.
func (m *Decoder) ToBundle() (*bundle.Bundle, error) {
	d := &bundle.Bundle{}
	if err := m.decode(d); err != nil {
		return nil, fmt.Errorf("while converting to Bundle, %v", err)
	}
	return d, nil
}

// ToBundleBuilder converts input data to the BundleBuilder type.
func (m *Decoder) ToBundleBuilder() (*bundle.BundleBuilder, error) {
	d := &bundle.BundleBuilder{}
	if err := m.decode(d); err != nil {
		return nil, fmt.Errorf("while converting to BundleBuilder, %v", err)
	}
	return d, nil
}

// ToComponent converts input data to the Component type.
func (m *Decoder) ToComponent() (*bundle.Component, error) {
	d := &bundle.Component{}
	if err := m.decode(d); err != nil {
		return nil, fmt.Errorf("while converting to Component, %v", err)
	}
	return d, nil
}

// ToComponentBuilder converts input data to the ComponentBuilder type.
func (m *Decoder) ToComponentBuilder() (*bundle.ComponentBuilder, error) {
	d := &bundle.ComponentBuilder{}
	if err := m.decode(d); err != nil {
		return nil, fmt.Errorf("while converting to ComponentBuilder, %v", err)
	}
	return d, nil
}

// ToComponentSet converts input data to the ComponentSet type.
func (m *Decoder) ToComponentSet() (*bundle.ComponentSet, error) {
	d := &bundle.ComponentSet{}
	if err := m.decode(d); err != nil {
		return nil, fmt.Errorf("while converting to ComponentSet, %v", err)
	}
	return d, nil
}

// ToUnstructured converts input data to the Unstructured type.
func (m *Decoder) ToUnstructured() (*unstructured.Unstructured, error) {
	d := &unstructured.Unstructured{}
	if err := m.decode(d); err != nil {
		return nil, fmt.Errorf("while converting to Unstructured, %v", err)
	}
	return d, nil
}

// ToJSONMap converts from json/yaml data to a map of string-to-interface.
func (m *Decoder) ToJSONMap() (map[string]interface{}, error) {
	d := make(map[string]interface{})
	if err := m.decode(&d); err != nil {
		return nil, fmt.Errorf("while converting to JSON Map, %v", err)
	}
	return d, nil
}

// ToObject converts to an arbitrary object via standard YAML / JSON
// serialization.
func (m *Decoder) ToObject(obj interface{}) error {
	return m.decode(obj)
}
