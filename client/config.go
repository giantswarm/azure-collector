package client

import (
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/giantswarm/microerror"
)

type AzureClientSetConfig struct {
	// ClientID is the ID of the Active Directory Service Principal.
	ClientID string
	// ClientSecret is the secret of the Active Directory Service Principal.
	ClientSecret string
	// EnvironmentName is the cloud environment identifier on Azure. Values can be
	// used as listed in the link below.
	//
	//     https://github.com/Azure/go-autorest/blob/ec5f4903f77ed9927ac95b19ab8e44ada64c1356/autorest/azure/environments.go#L13
	//
	EnvironmentName string
	// SubscriptionID is the ID of the Azure subscription.
	SubscriptionID string
	// TenantID is the ID of the Active Directory tenant.
	TenantID string
	// PartnerID is the ID used for the Azure Partner Program.
	PartnerID string
}

const (
	defaultAzureEnvironment = "AZUREPUBLICCLOUD"
	defaultAzureGUID        = "37f13270-5c7a-56ff-9211-8426baaeaabd"
)

// NewAzureClientSetConfig creates a new azure client set config and applies defaults.
func NewAzureClientSetConfig(clientid, clientsecret, subscriptionid, tenantid, environmentname, partnerid string) (AzureClientSetConfig, error) {
	if clientid == "" {
		return AzureClientSetConfig{}, microerror.Maskf(invalidConfigError, "ClientID must not be empty")
	}
	if clientsecret == "" {
		return AzureClientSetConfig{}, microerror.Maskf(invalidConfigError, "ClientSecret must not be empty")
	}
	if subscriptionid == "" {
		return AzureClientSetConfig{}, microerror.Maskf(invalidConfigError, "SubscriptionID must not be empty")
	}
	if tenantid == "" {
		return AzureClientSetConfig{}, microerror.Maskf(invalidConfigError, "TenantID must not be empty")
	}

	if environmentname == "" {
		environmentname = defaultAzureEnvironment
	}

	// No having partnerID in the secret means that customer has not
	// upgraded yet to use the Azure Partner Program. In that case we set a
	// constant random generated GUID that we haven't registered with Azure.
	// When all customers have migrated, we should error out instead.
	if partnerid == "" {
		partnerid = defaultAzureGUID
	}

	return AzureClientSetConfig{
		ClientID:        clientid,
		ClientSecret:    clientsecret,
		EnvironmentName: environmentname,
		PartnerID:       partnerid,
		SubscriptionID:  subscriptionid,
		TenantID:        tenantid,
	}, nil
}

// clientConfig contains all essential information to create an Azure client.
type clientConfig struct {
	subscriptionID          string
	partnerIdUserAgent      string
	resourceManagerEndpoint string
	servicePrincipalToken   *adal.ServicePrincipalToken
}
