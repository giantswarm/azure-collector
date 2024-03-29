package credential

import (
	"context"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v3/client"
	"github.com/giantswarm/azure-collector/v3/service/collector/key"
)

const (
	ClientIDKey       = "azure.azureoperator.clientid"
	ClientSecretKey   = "azure.azureoperator.clientsecret"
	SubscriptionIDKey = "azure.azureoperator.subscriptionid"
	TenantIDKey       = "azure.azureoperator.tenantid"
	PartnerIDKey      = "azure.azureoperator.partnerid"
	SingleTenantSP    = "giantswarm.io/single-tenant-service-principal"
)

func GetAzureConfigFromSecretName(ctx context.Context, ctrlClient ctrlclient.Client, name, namespace, gsTenantID string) (*client.AzureClientSetConfig, error) {
	credential := &v1.Secret{}
	err := ctrlClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, credential)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return GetAzureConfigFromSecret(credential, gsTenantID)
}

func GetAzureConfigFromSecret(credential *v1.Secret, gsTenantID string) (*client.AzureClientSetConfig, error) {
	clientID, err := valueFromSecret(credential, ClientIDKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clientSecret, err := valueFromSecret(credential, ClientSecretKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	subscriptionID, err := valueFromSecret(credential, SubscriptionIDKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantID, err := valueFromSecret(credential, TenantIDKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	partnerID, err := valueFromSecret(credential, PartnerIDKey)
	if err != nil {
		partnerID = ""
	}

	// By default we assume that the tenant cluster resources will belong to a subscription that belongs to a different Tenant ID than the one used for authentication.
	// Typically this means we are using a Service Principal from the GiantSwarm Tenant ID.
	credentials := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID)
	credentials.AuxTenants = append(credentials.AuxTenants, gsTenantID)
	if _, exists := credential.GetLabels()[SingleTenantSP]; exists || tenantID == gsTenantID {
		// In this case the tenant cluster resources will belong to a subscription that belongs to the same Tenant ID used for authentication.
		// Typically this means we are using a Service Principal from the customer Tenant ID.
		credentials = auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID)
	}

	authorizer, err := credentials.Authorizer()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	azureClientSetConfig, err := client.NewAzureClientSetConfig(
		authorizer,
		clientID,
		clientSecret,
		subscriptionID,
		partnerID,
		tenantID,
		gsTenantID,
	)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &azureClientSetConfig, nil
}

func GetAzureClientSetsFromCredentialSecrets(ctx context.Context, ctrlClient ctrlclient.Client, gsTenantID string) (map[*client.AzureClientSetConfig]*client.AzureClientSet, error) {
	azureClientSets := map[*client.AzureClientSetConfig]*client.AzureClientSet{}

	secrets, err := GetCredentialSecrets(ctx, ctrlClient)
	if err != nil {
		return azureClientSets, microerror.Mask(err)
	}

	for _, secret := range secrets {
		azureClientSetConfig, err := GetAzureConfigFromSecret(&secret, gsTenantID) //nolint:gosec
		if err != nil {
			return azureClientSets, microerror.Mask(err)
		}

		clientSet, err := client.NewAzureClientSet(*azureClientSetConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		azureClientSets[azureClientSetConfig] = clientSet
	}

	return azureClientSets, nil
}

func GetAzureClientSetsFromCredentialSecretsBySubscription(ctx context.Context, ctrlClient ctrlclient.Client, gsTenantID string) (map[string]*client.AzureClientSet, error) {
	azureClientSets := map[string]*client.AzureClientSet{}

	rawAzureClientSets, err := GetAzureClientSetsFromCredentialSecrets(ctx, ctrlClient, gsTenantID)
	if err != nil {
		return azureClientSets, microerror.Mask(err)
	}

	for azureClientSetConfig, azureClientSet := range rawAzureClientSets {
		azureClientSets[azureClientSetConfig.SubscriptionID] = azureClientSet
	}

	return azureClientSets, nil
}

func GetAzureClientSetsByCluster(ctx context.Context, ctrlClient ctrlclient.Client, gsTenantID string) (map[string]*client.AzureClientSet, error) {
	azureClientSets := map[string]*client.AzureClientSet{}
	var crs []providerv1alpha1.AzureConfig
	{
		mark := ""
		page := 0
		for page == 0 || len(mark) > 0 {
			opts := ctrlclient.ListOptions{
				Continue: mark,
			}
			list := providerv1alpha1.AzureConfigList{}
			err := ctrlClient.List(ctx, &list, &opts)
			if err != nil {
				return azureClientSets, microerror.Mask(err)
			}

			crs = append(crs, list.Items...)

			mark = list.Continue
			page++
		}
	}

	for _, cr := range crs {
		config, err := GetAzureConfigFromSecretName(ctx, ctrlClient, key.CredentialName(cr), key.CredentialNamespace(cr), gsTenantID)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		azureClients, err := client.NewAzureClientSet(*config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		azureClientSets[cr.GetName()] = azureClients
	}

	return azureClientSets, nil
}

func GetCredentialSecrets(ctx context.Context, ctrlClient ctrlclient.Client) (secrets []v1.Secret, err error) {
	mark := ""
	page := 0
	for page == 0 || len(mark) > 0 {
		opts := ctrlclient.ListOptions{
			Continue: mark,
		}
		list := v1.SecretList{}
		err := ctrlClient.List(ctx, &list, &opts, ctrlclient.MatchingLabels{"app": "credentiald"})
		if err != nil {
			return secrets, microerror.Mask(err)
		}

		secrets = append(secrets, list.Items...)

		mark = list.Continue
		page++
	}

	return secrets, nil
}

func valueFromSecret(secret *v1.Secret, key string) (string, error) {
	v, ok := secret.Data[key]
	if !ok {
		return "", microerror.Maskf(missingValueError, key)
	}

	return string(v), nil
}
