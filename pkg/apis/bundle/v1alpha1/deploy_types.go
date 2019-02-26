package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PackageDeploymentClass represents an immutable, versioned, ordered list of
// component sets.
type PackageDeploymentClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ComponentSets []ComponentSetReference `json:"componentSets,omitempty"`
	Version       string                  `json:"version,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// SetDeploymentClassList contains a list of SetDeploymentClasses
type PackageDeploymentClassList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []PackageDeploymentClass `json:"items,omitempty"`
}

// PackageDeploymentSpec represents a desired state for a deployed component set.
type PackageDeploymentSpec struct {
	ComponentSets []ComponentSetReference `json:"componentSets,omitempty"`
	Version       string                  `json:"version,omitempty"`
}

// PackageDeployment represents the current state of a each component in a
// deployed component set
type PackageDeploymentStatus struct {
	// ComponentName is the readable name of a component.
	Components []ComponentStatus `json:"components,omitempty"`
}

// ComponentStatus tracks the deployment status of a specific component
type ComponentStatus struct {
	ComponentName string `json:"componentName,omitempty"`
	Version       string `json:"version,omitempty"`
	Status        string `json:"status,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PackageDeploymentList contains a list of PackageDeployments
type PackageDeploymentList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []PackageDeployment `json:"items,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PackageDeployment is used to track a deployment of a package of component sets.
type PackageDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageDeploymentSpec   `json:"spec,omitempty"`
	Status PackageDeploymentStatus `json:"status,omitempty"`
}
