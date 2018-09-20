package converter

import (
	"fmt"
	"regexp"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// KubeResourceMap creates a map from a Kubernetes resource object reference to
// the instance.
func KubeResourceMap(resources []map[string]interface{}) (map[core.ObjectReference]interface{}, error) {
	crMap := make(map[core.ObjectReference]interface{})
	for _, cr := range resources {
		ref, err := ObjectRefFromRawKubeResource(cr)
		if err != nil {
			return nil, fmt.Errorf("error creating Kubernetes resource map: %v", err)
		}
		crMap[ref] = cr
	}
	return crMap, nil
}

// ObjectRefFromRawKubeResource extracts the ObjectReference out of a Kubernetes resource.
// - returns an error if there is no apiVersion or kind field in the resource
// - returns an error if the apiVersion value is not of the format "group/version"
func ObjectRefFromRawKubeResource(cr map[string]interface{}) (core.ObjectReference, error) {
	nullResp := core.ObjectReference{}
	ref := core.ObjectReference{}
	apiVersion := cr["apiVersion"]
	if apiVersion == nil {
		return nullResp, fmt.Errorf("no apiVersion field was found for Kubernetes resource %v", cr)
	}
	matches := regexp.MustCompile("(.+)/(.+)").FindStringSubmatch(apiVersion.(string))
	// The number of matches should be 3 - FindSubstringMatch returns the full matched string in
	// addition to the matched subexpressions.
	if matches == nil || len(matches) != 3 {
		return nullResp, fmt.Errorf("Kubernetes resource apiVersion is not formatted as group/version: got %q", apiVersion)
	}
	ref.APIVersion = apiVersion.(string)
	kind := cr["kind"]
	if kind == nil {
		return nullResp, fmt.Errorf("no kind field was found for Kubernetes resource %v", cr)
	}
	ref.Kind = kind.(string)

	md := cr["metadata"]
	if md == nil {
		return nullResp, fmt.Errorf("no metadata field was found for Kubernetes resource %v", cr)
	}

	metadata := md.(map[string]interface{})
	name := metadata["name"]
	if name == nil {
		return nullResp, fmt.Errorf("no metadata.name field was found for Kubernetes resource %v", cr)
	}

	ref.Name = name.(string)
	return ref, nil
}
