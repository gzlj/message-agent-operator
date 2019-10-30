package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MessageAgentSpec defines the desired state of MessageAgent
// +k8s:openapi-gen=true
type MessageAgentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Image         string     `json:"image"`
	MessageCenter string     `json:"messageCenter"`
	ClientId      string     `json:"clientId"`
	ClientSecret  string     `json:"clientSecret"`
	ServerPort    string     `json:"serverPort"`
	ApplyMsgType  string     `json:"applyMsgType"`
	Channels      []string   `json:"channels"`
	Receivers     []Receiver `json:"receivers"`

	Size      *int32                      `json:"size"`
}

// MessageAgentStatus defines the observed state of MessageAgent
// +k8s:openapi-gen=true
type MessageAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`
	AvailableReplicas int32 `json:"availableReplicas,omitempty" protobuf:"varint,4,opt,name=availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MessageAgent is the Schema for the messageagents API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MessageAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MessageAgentSpec   `json:"spec,omitempty"`
	Status MessageAgentStatus `json:"status,omitempty"`
}

type Receiver struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Closable bool `json:"closable"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MessageAgentList contains a list of MessageAgent
type MessageAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MessageAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MessageAgent{}, &MessageAgentList{})
}
