package find

import (
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// ComponentImage is a helper struct for returning found images for cluster components.k
type ComponentImage struct {
	// Key represents the key for this component.
	Key core.ClusterObjectKey

	// Path within a cluster object to find the container image.
	ObjectPath []string

	// Image is the image name. Usually having the form
	// `<registry>/<repository>/<image>:<tag>`. For example:
	// `gcr.io/google_containers/etcd:3.1.11`
	Image string
}

// ComponentImages returns all the images from the components
func (b *BundleFinder) ComponentImages() ([]*ComponentImage, error) {
	var images []*ComponentImage
	for _, ca := range b.bundle.GetSpec().GetComponents() {
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
			partial, err := b.findImagesInKubeObj(obj)
			if err != nil {
				return nil, err
			}
			images = append(images, partial...)
		}
	}
	return images, nil
}

func (b *BundleFinder) findImagesInKubeObj(s *structpb.Struct) ([]*ComponentImage, error) {
	var images []*ComponentImage
	return images, nil
}
