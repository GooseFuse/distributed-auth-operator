package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AuthClusterSpec defines the desired state of AuthCluster
type AuthClusterSpec struct {
	NodeCount int    `json:"nodeCount"`
	RedisURL  string `json:"redisURL"`
}

// AuthClusterStatus defines the observed state of AuthCluster
type AuthClusterStatus struct {
	ReadyNodes int `json:"readyNodes"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type AuthCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthClusterSpec   `json:"spec,omitempty"`
	Status AuthClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type AuthClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthCluster{}, &AuthClusterList{})
}
