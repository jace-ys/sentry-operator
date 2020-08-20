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

// TeamSpec defines the desired state of Team.
type TeamSpec struct {
	// Name of the Sentry team.
	Name string `json:"name,omitempty"`

	// Slug of the Sentry team.
	Slug string `json:"slug,omitempty"`
}

// +kubebuilder:validation:Enum=Created;Error
type TeamCondition string

const (
	TeamConditionCreated TeamCondition = "Created"
	TeamConditionError   TeamCondition = "Error"
)

// TeamStatus defines the observed state of Team.
type TeamStatus struct {
	// The state of the Sentry team.
	// "Created" indicates that the Sentry team was created successfully.
	// "Error" indicates that an error occurred while trying to reconcile the Sentry team.
	Condition TeamCondition `json:"condition,omitempty"`

	// Additional detail about any errors that occurred while trying to reconcile the Sentry team.
	Message string `json:"message,omitempty"`

	// The ID of the Sentry team.
	ID string `json:"id,omitempty"`

	// The time that the Sentry team was last successfully reconciled.
	LastSynced *metav1.Time `json:"lastSynced,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.condition`

// Team is the Schema for the teams API.
type Team struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TeamSpec   `json:"spec,omitempty"`
	Status TeamStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TeamList contains a list of Team.
type TeamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Team `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Team{}, &TeamList{})
}
