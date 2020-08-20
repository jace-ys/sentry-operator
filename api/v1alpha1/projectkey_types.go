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

// ProjectKeySpec defines the desired state of ProjectKey.
type ProjectKeySpec struct {
	// Slug of the Sentry project that this project key should be created under.
	Project string `json:"project,omitempty"`

	// Name of the Sentry project key.
	Name string `json:"name,omitempty"`
}

// +kubebuilder:validation:Enum=Created;Error
type ProjectKeyCondition string

const (
	ProjectKeyConditionCreated ProjectKeyCondition = "Created"
	ProjectKeyConditionError   ProjectKeyCondition = "Error"
)

// ProjectKeyStatus defines the observed state of ProjectKey.
type ProjectKeyStatus struct {
	// The state of the Sentry project key.
	// "Created" indicates that the Sentry project key was created successfully.
	// "Error" indicates that an error occurred while trying to reconcile the Sentry project key.
	Condition ProjectKeyCondition `json:"condition,omitempty"`

	// Additional detail about any errors that occurred while trying to reconcile the Sentry project key.
	Message string `json:"message,omitempty"`

	// The ID of the Sentry project key.
	ID string `json:"id,omitempty"`

	// The time that the Sentry project key was last successfully reconciled.
	LastSynced *metav1.Time `json:"lastSynced,omitempty"`

	// The ID of the Sentry project that this project key belongs to.
	ProjectID string `json:"projectID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.condition`

// ProjectKey is the Schema for the projectkeys API.
type ProjectKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectKeySpec   `json:"spec,omitempty"`
	Status ProjectKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectKeyList contains a list of ProjectKey.
type ProjectKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectKey{}, &ProjectKeyList{})
}
