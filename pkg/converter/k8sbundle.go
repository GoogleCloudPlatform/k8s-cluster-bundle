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
	"github.com/ghodss/yaml"

	kbun "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/k8sbundle/v1alpha1"
)

// YAMLToKBundle converts from a YAML object to the k8sbundle object.
func YAMLToK8sBundle(b []byte) (*kbun.ClusterBundle, error) {
	bun := &kbun.ClusterBundle{}
	err := yaml.Unmarshal(b, bun)
	if err != nil {
		return nil, err
	}
	return bun, nil
}
