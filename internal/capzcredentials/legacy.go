package capzcredentials

import (
	"context"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	apiextensionslabels "github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	credentialDefaultNamespace = "giantswarm"
	credentialDefaultName      = "credential-default" // nolint:gosec

	clientIDKey       = "azure.azureoperator.clientid"
	clientSecretKey   = "azure.azureoperator.clientsecret"
	subscriptionIDKey = "azure.azureoperator.subscriptionid"
	tenantIDKey       = "azure.azureoperator.tenantid"
)

func getLegacyCredentials(ctx context.Context, ctrlClient client.Client, objectMeta metav1.ObjectMeta) (*AzureCredentials, error) {
	credentialSecret, err := getCredentialSecret(ctx, ctrlClient, objectMeta)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secret := &corev1.Secret{}
	err = ctrlClient.Get(ctx, client.ObjectKey{Namespace: credentialSecret.Namespace, Name: credentialSecret.Name}, secret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clientID, err := valueFromSecret(secret, clientIDKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clientSecret, err := valueFromSecret(secret, clientSecretKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantID, err := valueFromSecret(secret, tenantIDKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	subscriptionID, err := valueFromSecret(secret, subscriptionIDKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &AzureCredentials{
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
	}, nil

}

func getCredentialSecret(ctx context.Context, ctrlClient client.Client, objectMeta metav1.ObjectMeta) (*v1alpha1.CredentialSecret, error) {
	var err error
	var credentialSecret *v1alpha1.CredentialSecret

	credentialSecret, err = getOrganizationCredentialSecret(ctx, ctrlClient, objectMeta)
	if IsCredentialsNotFoundError(err) {
		credentialSecret, err = getLegacyCredentialSecret(ctx, ctrlClient, objectMeta)
		if IsCredentialsNotFoundError(err) {
			return &v1alpha1.CredentialSecret{
				Namespace: credentialDefaultNamespace,
				Name:      credentialDefaultName,
			}, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return credentialSecret, nil
}

// getOrganizationCredentialSecret tries to find a Secret in the organization namespace.
func getOrganizationCredentialSecret(ctx context.Context, ctrlClient client.Client, objectMeta metav1.ObjectMeta) (*v1alpha1.CredentialSecret, error) {
	secretList := &corev1.SecretList{}
	{
		err := ctrlClient.List(
			ctx,
			secretList,
			client.InNamespace(objectMeta.Namespace),
			client.MatchingLabels{
				apiextensionslabels.App:          "credentiald",
				apiextensionslabels.Organization: organizationID(objectMeta),
			},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We currently only support one credential secret per organization.
	// If there are more than one, return an error.
	if len(secretList.Items) > 1 {
		return nil, microerror.Mask(tooManyCredentialsError)
	}

	if len(secretList.Items) < 1 {
		return nil, microerror.Mask(credentialsNotFoundError)
	}

	// If one credential secret is found, we use that.
	secret := secretList.Items[0]

	credentialSecret := &v1alpha1.CredentialSecret{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	return credentialSecret, nil
}

// getLegacyCredentialSecret tries to find a Secret in the default credentials namespace but labeled with the organization name.
// This is needed while we migrate everything to the org namespace and org credentials are created in the org namespace instead of the default namespace.
func getLegacyCredentialSecret(ctx context.Context, ctrlClient client.Client, objectMeta metav1.ObjectMeta) (*v1alpha1.CredentialSecret, error) {
	secretList := &corev1.SecretList{}
	{
		err := ctrlClient.List(
			ctx,
			secretList,
			client.InNamespace(credentialDefaultNamespace),
			client.MatchingLabels{
				apiextensionslabels.App:          "credentiald",
				apiextensionslabels.Organization: organizationID(objectMeta),
			},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We currently only support one credential secret per organization.
	// If there are more than one, return an error.
	if len(secretList.Items) > 1 {
		return nil, microerror.Mask(tooManyCredentialsError)
	}

	if len(secretList.Items) < 1 {
		return nil, microerror.Mask(credentialsNotFoundError)
	}

	// If one credential secret is found, we use that.
	secret := secretList.Items[0]

	credentialSecret := &v1alpha1.CredentialSecret{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	return credentialSecret, nil
}

func organizationID(getter metav1.ObjectMeta) string {
	return getter.GetLabels()[apiextensionslabels.Organization]
}

func valueFromSecret(secret *corev1.Secret, key string) (string, error) {
	v, ok := secret.Data[key]
	if !ok {
		return "", microerror.Maskf(missingValueError, key)
	}

	return string(v), nil
}
