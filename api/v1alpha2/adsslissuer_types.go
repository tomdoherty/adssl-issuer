/*
Copyright (c) 2020 Tom Doherty <tom@tomdoherty.io>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIEDi
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretKeySelector contains the reference to a secret.
type SecretKeySelector struct {
	// The name of the secret in the pod's namespace to select from.
	Name string `json:"name"`

	// The key of the secret to select from. Must be a valid secret key.
	// +optional
	Key string `json:"key,omitempty"`
}

// AdsslIssuerSpec defines the desired state of AdsslIssuer
type AdsslIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URL is the base URL for the AD server
	URL string `json:"url"`

	// User is the username to authenticate with
	User string `json:"user"`

	// PasswordRef is a reference to a Secret containing the provisioner
	// password used to decrypt the provisioner private key.
	PasswordRef SecretKeySelector `json:"passwordRef"`
}

// AdsslIssuerStatus defines the observed state of AdsslIssuer
type AdsslIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	Conditions []AdsslIssuerCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// AdsslIssuer is the Schema for the adsslissuers API
type AdsslIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdsslIssuerSpec   `json:"spec,omitempty"`
	Status AdsslIssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AdsslIssuerList contains a list of AdsslIssuer
type AdsslIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdsslIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdsslIssuer{}, &AdsslIssuerList{})
}

// ConditionType represents a AdsslIssuer condition type.
// +kubebuilder:validation:Enum=Ready
type ConditionType string

const (
	// ConditionReady indicates that a AdsslIssuer is ready for use.
	ConditionReady ConditionType = "Ready"
)

// ConditionStatus represents a condition's status.
// +kubebuilder:validation:Enum=True;False;Unknown
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// AdsslIssuerCondition contains condition information for the step issuer.
type AdsslIssuerCondition struct {
	// Type of the condition, currently ('Ready').
	Type ConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}
