package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Model defines the details of the model to validate.
type Model struct {
	Path          string `json:"path"`
	SignaturePath string `json:"signaturePath"`
}

type SigstoreConfig struct {
	CertificateIdentity   string `json:"certificateIdentity,omitempty"`
	CertificateOidcIssuer string `json:"certificateOidcIssuer,omitempty"`
}

type PkiConfig struct {
	// Path to the certificate authority for PKI.
	CertificateAuthority string `json:"certificateAuthority,omitempty"`
}

type PrivateKeyConfig struct {
	// Path to the private key.
	KeyPath string `json:"keyPath,omitempty"`
}

type ValidationConfig struct {
	SigstoreConfig   *SigstoreConfig   `json:"sigstoreConfig,omitempty"`
	PkiConfig        *PkiConfig        `json:"pkiConfig,omitempty"`
	PrivateKeyConfig *PrivateKeyConfig `json:"privateKeyConfig,omitempty"`
}

// ModelValidationSpec defines the desired state of ModelValidation.
type ModelValidationSpec struct {
	// Model details.
	Model Model `json:"model"`
	// Configuration for validation methods.
	Config ValidationConfig `json:"config"`
}

// ModelValidationStatus defines the observed state of ModelValidation.
type ModelValidationStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ModelValidation is the Schema for the modelvalidations API
type ModelValidation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelValidationSpec   `json:"spec,omitempty"`
	Status ModelValidationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModelValidationList contains a list of ModelValidation
type ModelValidationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelValidation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelValidation{}, &ModelValidationList{})
}
