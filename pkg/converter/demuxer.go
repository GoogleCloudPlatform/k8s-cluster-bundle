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
	"strings"

	"github.com/ghodss/yaml"
)

// FromObject creates a converting demuxer from an object.
func FromObject(obj interface{}) *Demuxer {
	return &Demuxer{obj}
}

// Demuxer converts from an object to some serialized format for the object.
type Demuxer struct {
	obj interface{}
}

func (m *Demuxer) demux(format ContentType) ([]byte, error) {
	switch format {
	case YAML:
		return yaml.Marshal(m.obj)
	case JSON:
		return json.Marshal(m.obj)
	default:
		return nil, fmt.Errorf("unknown content type: %q", format)
	}
}

// ToYAML converts an object to YAML
func (m *Demuxer) ToYAML() ([]byte, error) {
	return m.demux(YAML)
}

// ToJSON converts an object to JSON
func (m *Demuxer) ToJSON() ([]byte, error) {
	return m.demux(JSON)
}

// ToContentType converts to a custom content type.
func (m *Demuxer) ToContentType(ctype string) ([]byte, error) {
	lower := strings.ToLower(ctype)
	return m.demux(ContentType(lower))
}
