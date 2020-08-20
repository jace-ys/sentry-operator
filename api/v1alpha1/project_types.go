/*

MIT License

Copyright (c) 2020 Jace Tan

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectSpec defines the desired state of Project.
type ProjectSpec struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=50
	// Slug of the Sentry team that this project should be created under.
	Team string `json:"team"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=50
	// Name of the Sentry project.
	Name string `json:"name"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=50
	// Slug of the Sentry project.
	Slug string `json:"slug"`
}

// +kubebuilder:validation:Enum=Created;Error
type ProjectCondition string

const (
	ProjectConditionCreated ProjectCondition = "Created"
	ProjectConditionError   ProjectCondition = "Error"
)

// ProjectStatus defines the observed state of Project.
type ProjectStatus struct {
	// The state of the Sentry project.
	// "Created" indicates that the Sentry project was created successfully.
	// "Error" indicates that an error occurred while trying to reconcile the Sentry project.
	Condition ProjectCondition `json:"condition,omitempty"`

	// Additional detail about any errors that occurred while trying to reconcile the Sentry project.
	Message string `json:"message,omitempty"`

	// The ID of the Sentry project.
	ID string `json:"id,omitempty"`

	// The time that the Sentry project was last successfully reconciled.
	LastSynced *metav1.Time `json:"lastSynced,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.condition`

// Project is the Schema for the projects API.
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project.
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
