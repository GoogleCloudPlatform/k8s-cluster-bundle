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

package build

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
)

func BenchmarkBuildAndInline_Component(t *testing.B) {
	b, err := ioutil.ReadFile("../../examples/component/etcd-component-builder.yaml")
	if err != nil {
		t.Fatal(err)
	}

	dataPath := "../../examples/component/etcd-component-builder.yaml"

	for i := 0; i < t.N; i++ {
		cb, err := converter.FromYAML(b).ToComponentBuilder()
		if err != nil {
			t.Fatal(err)
		}
		inliner := NewLocalInliner("../../examples/component/")
		component, err := inliner.ComponentFiles(context.Background(), cb, dataPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = converter.FromObject(component).ToYAML()
		if err != nil {
			t.Fatal(err)
		}
		validate.Component(component)
	}

}
