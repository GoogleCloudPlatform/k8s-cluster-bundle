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

package find

import (
	"fmt"
	"strings"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// ImageFinder finds container and OS Images in Bundles.
type ImageFinder struct {
	Bundle *bpb.ClusterBundle
}

// ContainerImage is a helper struct for returning found container images for cluster objects.
type ContainerImage struct {
	// Key represents the key for representing the specific cluster object that
	// this is from.
	Key core.ClusterObjectKey

	// Image are the images used by the cluster object. Usually having the form
	// `<registry>/<repository>/<image>:<tag>`. For example:
	// `gcr.io/google_containers/etcd:3.1.11`
	Image string
}

// String converts the ContainerImage into a human-readable string.
func (c *ContainerImage) String() string {
	return fmt.Sprintf("{Key:%v, Image:%q}", c.Key, c.Image)
}

// ContainerImages returns all the images from the cluster components in a Bundle.
func (b *ImageFinder) AllContainerImages() []*ContainerImage {
	var images []*ContainerImage
	b.WalkAllContainerImages(func(key core.ClusterObjectKey, img string) string {
		images = append(images, &ContainerImage{key, img})
		return img
	})
	return images
}

// ContainerImages returns all the images from a single Kubernetes object.
func (b *ImageFinder) ContainerImages(key core.ClusterObjectKey, st *structpb.Struct) []*ContainerImage {
	var images []*ContainerImage
	b.WalkContainerImages(st, func(img string) string {
		images = append(images, &ContainerImage{key, img})
		return img
	})
	return images
}

// WalkContainerImages provides a method for traversing through Container
// images in a single kubernetes object, with a function for processing each
// container value.
//
// If an image value is returned from the function that is not equal to the input
// value, the value is replaced with the new value.
//
// This changes the Bundle object in-place, so if changes are intended, it is
// recommend that the Bundle be cloned.
func (b *ImageFinder) WalkContainerImages(st *structpb.Struct, fn func(img string) string) {
	// It would be more robust to just be aware of Pods, Deployments, and the
	// various K8S types that have container images rather then recursing through
	// everything.  It's possible, for example, that we that we might encouncer
	// an 'image' field in some options custom resource that's unintended.
	containerImageRecurser("", "", &structpb.Value{
		Kind: &structpb.Value_StructValue{st},
	}, fn)
}

// WalkAllContainerImages works the same as WalkContainerImages, except all
// images are traversed. Additionally, the cluster object context is also
// provided. Note that objects must be inlined to be walked.
//
// This changes the Bundle object in-place, so if changes are intended, it is
// recommend that the Bundle be cloned.
func (b *ImageFinder) WalkAllContainerImages(fn func(key core.ClusterObjectKey, img string) string) {
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		compName := ca.GetMetadata().GetName()
		for _, obj := range ca.GetSpec().GetClusterObjects() {
			objName := core.ObjectName(obj)
			if obj == nil {
				continue
			}
			key := core.ClusterObjectKey{
				ComponentName: compName,
				ObjectName:    objName,
			}
			b.WalkContainerImages(obj, func(img string) string {
				return fn(key, img)
			})
		}
	}
}

// ContainerImageRecurser is a function that looks through a struct pb for
// fields named "Image" and calls a function on the resulting value.
//
// If an image value is returned from the function and the value is not equal
// to the input value.
func containerImageRecurser(fieldName string, parentFieldName string, st *structpb.Value, fn func(img string) string) {
	switch st.Kind.(type) {
	case *structpb.Value_NullValue:
	case *structpb.Value_NumberValue:
	case *structpb.Value_StringValue:
		// From my spotty research, it's almost always true that the parent name
		// for the container object is 'container', 'containers' or
		// 'somethingContainer[s]'.
		if fieldName == "image" && (strings.Contains(parentFieldName, "container") || strings.Contains(parentFieldName, "Container")) ||
			fieldName == "url" && parentFieldName == "osImage" { // hack to make finding images work with NodeConfigs
			if ret := fn(st.GetStringValue()); ret != st.GetStringValue() {
				st.Kind = &structpb.Value_StringValue{ret}
			}
		}
	case *structpb.Value_BoolValue:
	case *structpb.Value_StructValue:
		for k, v := range st.GetStructValue().GetFields() {
			// Swap parentFieldName with fieldName
			containerImageRecurser(k, fieldName, v, fn)
		}
	case *structpb.Value_ListValue:
		for _, val := range st.GetListValue().GetValues() {
			// Preserve the fieldname for the parent list object.
			containerImageRecurser(fieldName, parentFieldName, val, fn)
		}
	case nil:
	default:
		// Shouldn't happen. But if it does, move on.
	}
}

// WalkAllImages walks all node and container images. Only one of
// nodeConfigName or key will be filled out, based on whether the image is from
// a node config or from a cluster object.
//
// This changes the Bundle object in-place, so if changes are intended, it is
// recommend that the Bundle be cloned.
func (b *ImageFinder) WalkAllImages(fn func(key core.ClusterObjectKey, img string) string) {
	b.WalkAllContainerImages(fn)
}

// AllImages returns all images found -- both container images and OS images for nodes.
type AllImages struct {
	ContainerImages []*ContainerImage
}

// FindAllImages finds both container and node images.
func (b *ImageFinder) AllImages() *AllImages {
	return &AllImages{
		ContainerImages: b.AllContainerImages(),
	}
}

// Flattened turns an AllImages struct with image information into a struct
// containing lists of strings. All duplicates are removed.
func (a *AllImages) Flattened() *AllImagesFlattened {
	seen := make(map[string]bool)
	var containerImages []string
	for _, val := range a.ContainerImages {
		if !seen[val.Image] {
			containerImages = append(containerImages, val.Image)
		}
		seen[val.Image] = true
	}
	return &AllImagesFlattened{
		ContainerImages: containerImages,
	}
}

// ImagesFlattened contains images found, but flattened into lists of
// strings.
type AllImagesFlattened struct {
	ContainerImages []string `json:"containerImages"`
}
