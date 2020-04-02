package key

import (
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

// ClusterID returns the unique ID for this cluster.
func ClusterID(customObject providerv1alpha1.AzureConfig) string {
	return customObject.Spec.Cluster.ID
}

// CredentialName returns name of the credential secret.
func CredentialName(customObject providerv1alpha1.AzureConfig) string {
	return customObject.Spec.Azure.CredentialSecret.Name
}

// CredentialNamespace returns namespace of the credential secret.
func CredentialNamespace(customObject providerv1alpha1.AzureConfig) string {
	return customObject.Spec.Azure.CredentialSecret.Namespace
}
