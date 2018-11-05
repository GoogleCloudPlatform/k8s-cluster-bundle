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

package transformer

import (
	"strings"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

// ImageTransformer makes modifications to container and node images in
// Bundles.
type ImageTransformer struct {
	components []*bundle.ComponentPackage
}

// NewImageTransformer creates a new ImageTransformer instance.
func NewImageTransformer(comp []*bundle.ComponentPackage) *ImageTransformer {
	return &ImageTransformer{comp}
}

// ImageSubRule represents string substitutions to preform on images in
// bundles.
type ImageSubRule struct {
	Find    string
	Replace string
}

// TransformImagesStringSub transforms container and node images based on
// string substitution. A cloned bundle is always returned, even if no changes
// are made.
//
// Rules are applied in order to images. If two rules apply, then they will be
// applied in order.
func (t *ImageTransformer) TransformImagesStringSub(rules []*ImageSubRule) []*bundle.ComponentPackage {
	newComp := (&core.ComponentData{Components: t.components}).DeepCopy().Components
	finder := find.NewImageFinder(newComp)
	finder.WalkAllImages(func(_ core.ClusterObjectKey, img string) string {
		for _, r := range rules {
			if strings.Contains(img, r.Find) {
				// We can't return because it's possible that another rule will apply
				// later.
				img = strings.Replace(img, r.Find, r.Replace, 1)
			}
		}
		return img
	})
	return newComp
}
