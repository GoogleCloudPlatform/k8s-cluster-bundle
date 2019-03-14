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

package patchtmpl

import (
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	crdext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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

	must(appsv1.AddToScheme(k.KubeScheme))
	must(appsv1beta1.AddToScheme(k.KubeScheme))
	must(corev1.AddToScheme(k.KubeScheme))
	must(crdext.AddToScheme(k.KubeScheme))
	must(extv1beta1.AddToScheme(k.KubeScheme))
	must(policyv1beta1.AddToScheme(k.KubeScheme))
	must(rbacv1.AddToScheme(k.KubeScheme))
	must(storagev1.AddToScheme(k.KubeScheme))

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
