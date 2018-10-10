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

package scheme

import (
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	// apiextensions is a very heavy library to depend on. It's not used right
	// now anyway, so for now, we remove it.
	// apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func init() {
	s := runtime.NewScheme()
	k := &PatcherScheme{
		KubeScheme: s,
		Codecs:     serializer.NewCodecFactory(s),
	}

	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	must(corev1.AddToScheme(k.KubeScheme))
	must(extv1beta1.AddToScheme(k.KubeScheme))
	// must(apiextv1beta1.AddToScheme(k.KubeScheme))

	defaultPatcherScheme = k
}

// PatcherScheme wraps Kubernetes schema constructs.
type PatcherScheme struct {
	KubeScheme *runtime.Scheme
	Codecs     serializer.CodecFactory
}

var defaultPatcherScheme *PatcherScheme

// DefaultPatcherScheme returns a default scheme with several types built-in.
func DefaultPatcherScheme() *PatcherScheme {
	return defaultPatcherScheme
}
