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

// ProjectKeySpec defines the desired state of ProjectKey
type ProjectKeySpec struct {
	Name string `json:"name,omitempty"`
}

type ProjectKeyCondition string

const (
	ProjectKeyConditionCreated ProjectKeyCondition = "Created"
	ProjectKeyConditionError   ProjectKeyCondition = "Error"
)

// ProjectKeyStatus defines the observed state of ProjectKey
type ProjectKeyStatus struct {
	Condition ProjectCondition `json:"condition,omitempty"`
	Message   string           `json:"message,omitempty"`

	ID         string       `json:"id,omitempty"`
	LastSynced *metav1.Time `json:"lastSynced,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.condition`

// ProjectKey is the Schema for the projectkeys API
type ProjectKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectKeySpec   `json:"spec,omitempty"`
	Status ProjectKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectKeyList contains a list of ProjectKey
type ProjectKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectKey{}, &ProjectKeyList{})
}
