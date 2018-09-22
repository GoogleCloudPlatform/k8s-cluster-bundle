package find

import (
	"fmt"
	"strings"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// ImageFinder finds Images in Bundles.
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

func (c *ContainerImage) String() string {
	return fmt.Sprintf("{Key:%q, Image:%q}", c.Key, c.Image)
}

// ContainerImages returns all the images from the cluster components in a Bundle.
func (b *ImageFinder) ContainerImages() ([]*ContainerImage, error) {
	var images []*ContainerImage
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		compName := ca.GetName()
		if compName == "" {
			return nil, fmt.Errorf("cluster components must always have a name. was empty for %v", ca)
		}

		for _, co := range ca.GetClusterObjects() {
			objName := co.GetName()
			if objName == "" {
				return nil, fmt.Errorf("cluster component objects must always have a name. was empty for object %v in component %q", co, compName)
			}
			obj := co.GetInlined()
			if obj == nil {
				continue
			}
			key := core.ClusterObjectKey{
				ComponentName: compName,
				ObjectName:    objName,
			}
			if found := findImagesInKubeObj(key, obj); len(found) > 0 {
				images = append(images, found...)
			}
		}
	}
	return images, nil
}

func findImagesInKubeObj(key core.ClusterObjectKey, s *structpb.Struct) []*ContainerImage {
	var images []*ContainerImage
	// It would be more robust to just be aware of Pods, Deployments, and the
	// various K8S types that have container images rather then recursing through
	// everything.  It's possible, for example, that we that we might encouncer
	// an 'image' field in some options custom resource that's unintended.
	ImageRecurser("", "", &structpb.Value{
		Kind: &structpb.Value_StructValue{s},
	}, func(val *structpb.Value) {
		images = append(images, &ContainerImage{
			Key:   key,
			Image: val.GetStringValue(),
		})
	})

	return images
}

// ImageRecurser is a function that looks through a struct pb for fields named
// "Image" and calls a function on the resulting value.
func ImageRecurser(fieldName string, parentFieldName string, st *structpb.Value, fn func(*structpb.Value)) {
	switch st.Kind.(type) {
	case *structpb.Value_NullValue:
	case *structpb.Value_NumberValue:
	case *structpb.Value_StringValue:
		// From my spotty research, it's almost always true that the parent name
		// for the container object is 'container', 'containers' or
		// 'somethingContainer[s]'.
		if fieldName == "image" && (strings.Contains(parentFieldName, "container") || strings.Contains(parentFieldName, "Container")) {
			fn(st)
		}
	case *structpb.Value_BoolValue:
	case *structpb.Value_StructValue:
		for k, v := range st.GetStructValue().GetFields() {
			// Swap parentFieldName with fieldName
			ImageRecurser(k, fieldName, v, fn)
		}
	case *structpb.Value_ListValue:
		for _, val := range st.GetListValue().GetValues() {
			// Preserve the fieldname for the parent list object.
			ImageRecurser(fieldName, parentFieldName, val, fn)
		}
	case nil:
	default:
		// Shouldn't happen. But if it does, move on.
	}
}

// NodeImage represents an OS image coming from NodeConfig.
type NodeImage struct {
	ConfigName string
	Image      string
}

// NodeImages returns all the node images
func (b *ImageFinder) NodeImages() ([]*NodeImage, error) {
	var images []*NodeImage
	for _, config := range b.Bundle.GetSpec().GetNodeConfigs() {
		configName := config.GetName()
		if configName == "" {
			return nil, fmt.Errorf("node configs must always have a name. was empty for %v", config)
		}

		if url := config.GetOsImage().GetUrl(); url != "" {
			images = append(images, &NodeImage{configName, url})
		}
	}
	return images, nil
}

// AllImages returns all images found -- both container images and OS images for nodes.
type AllImages struct {
	NodeImages      []*NodeImage
	ContainerImages []*ContainerImage
}

// FindAllImages finds both container and node images.
func (b *ImageFinder) AllImages() (*AllImages, error) {
	ni, err := b.NodeImages()
	if err != nil {
		return nil, err
	}
	ci, err := b.ContainerImages()
	if err != nil {
		return nil, err
	}

	return &AllImages{
		NodeImages:      ni,
		ContainerImages: ci,
	}, nil
}

// Flattened turns an AllImages struct with image information into a struct
// containing lists of strings. All duplicates are removed.
func (a *AllImages) Flattened() *AllImagesFlattened {
	var nodeImages []string
	seen := make(map[string]bool)
	for _, val := range a.NodeImages {
		if !seen[val.Image] {
			nodeImages = append(nodeImages, val.Image)
		}
		seen[val.Image] = true
	}

	var containerImages []string
	seen = make(map[string]bool)
	for _, val := range a.ContainerImages {
		if !seen[val.Image] {
			containerImages = append(containerImages, val.Image)
		}
		seen[val.Image] = true
	}
	return &AllImagesFlattened{
		NodeImages:      nodeImages,
		ContainerImages: containerImages,
	}
}

// AllImagesFlattened returns all images found, but flattened into lists of strings.
type AllImagesFlattened struct {
	NodeImages      []string
	ContainerImages []string
}
