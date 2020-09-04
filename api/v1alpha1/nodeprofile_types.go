/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type Labels map[string]string

type NodeLabels struct {
	Required  Labels `json:"required,omitempty"`
	Preferred Labels `json:"preferred,omitempty"`
}

// NodeProfileSpec defines the desired state of NodeProfile
type NodeProfileSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NodeProfile. Edit NodeProfile_types.go to remove/update
	Labels NodeLabels `json:"labels,omitempty"`
}

// NodeProfileStatus defines the observed state of NodeProfile
type NodeProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	NodeNames []string `json:"nodeNames,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NodeProfile is the Schema for the nodeprofiles API
type NodeProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeProfileSpec   `json:"spec,omitempty"`
	Status NodeProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeProfileList contains a list of NodeProfile
type NodeProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeProfile{}, &NodeProfileList{})
}