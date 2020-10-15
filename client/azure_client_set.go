package client

import (
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/giantswarm/microerror"
)

type AzureClientSetConfig struct {
	Authorizer     autorest.Authorizer
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	PartnerID      string
	TenantID       string
	GSTenantID     string
}

const (
	defaultAzureGUID = "37f13270-5c7a-56ff-9211-8426baaeaabd"
)

// AzureClientSet is the collection of Azure API clients.
type AzureClientSet struct {
	ApplicationsClient *graphrbac.ApplicationsClient
	// DeploymentsClient manages deployments of ARM templates.
	DeploymentsClient *resources.DeploymentsClient
	// GroupsClient manages ARM resource groups.
	GroupsClient *resources.GroupsClient
	// UsageClient is used to work with limits and quotas.
	UsageClient *compute.UsageClient
	// VirtualNetworkGatewayConnectionsClient manages virtual network gateway connections.
	VirtualNetworkGatewayConnectionsClient *network.VirtualNetworkGatewayConnectionsClient
	// VirtualMachineScaleSetVMsClient manages virtual machine scale set VMs.
	VirtualMachineScaleSetVMsClient *compute.VirtualMachineScaleSetVMsClient
}

func init() {
	// ONE DOES NOT SIMPLY RETRY ON HTTP 429.
	autorest.StatusCodesForRetry = removeElementFromSlice(autorest.StatusCodesForRetry, http.StatusTooManyRequests)
}

// NewAzureClientSetConfig creates a new azure client set config and applies defaults.
func NewAzureClientSetConfig(authorizer autorest.Authorizer, clientid, clientsecret, subscriptionID, partnerID, tenantID, gsTenantID string) (AzureClientSetConfig, error) {
	// No having partnerID in the secret means that customer has not
	// upgraded yet to use the Azure Partner Program. In that case we set a
	// constant random generated GUID that we haven't registered with Azure.
	// When all customers have migrated, we should error out instead.
	if partnerID == "" {
		partnerID = defaultAzureGUID
	}

	return AzureClientSetConfig{
		Authorizer:     authorizer,
		ClientID:       clientid,
		ClientSecret:   clientsecret,
		PartnerID:      fmt.Sprintf("pid-%s", partnerID),
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,
		GSTenantID:     gsTenantID,
	}, nil
}

// NewAzureClientSet returns the Azure API clients.
func NewAzureClientSet(config AzureClientSetConfig) (*AzureClientSet, error) {
	applicationsClient, err := newApplicationsClient(config.ClientID, config.ClientSecret, config.GSTenantID, config.PartnerID)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	deploymentsClient, err := newDeploymentsClient(config.Authorizer, config.SubscriptionID, config.PartnerID)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	groupsClient, err := newGroupsClient(config.Authorizer, config.SubscriptionID, config.PartnerID)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	usageClient, err := newUsageClient(config.Authorizer, config.SubscriptionID, config.PartnerID)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	virtualNetworkGatewayConnectionsClient, err := newVirtualNetworkGatewayConnectionsClient(config.Authorizer, config.SubscriptionID, config.PartnerID)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	virtualMachineScaleSetVMsClient, err := newVirtualMachineScaleSetVMsClient(config.Authorizer, config.SubscriptionID, config.PartnerID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clientSet := &AzureClientSet{
		ApplicationsClient:                     applicationsClient,
		DeploymentsClient:                      deploymentsClient,
		GroupsClient:                           groupsClient,
		UsageClient:                            usageClient,
		VirtualNetworkGatewayConnectionsClient: virtualNetworkGatewayConnectionsClient,
		VirtualMachineScaleSetVMsClient:        virtualMachineScaleSetVMsClient,
	}

	return clientSet, nil
}

func prepareClient(client *autorest.Client, authorizer autorest.Authorizer, partnerID string) *autorest.Client {
	client.Authorizer = authorizer
	_ = client.AddToUserAgent(partnerID)

	return client
}

func newDeploymentsClient(authorizer autorest.Authorizer, subscriptionID, partnerID string) (*resources.DeploymentsClient, error) {
	client := resources.NewDeploymentsClient(subscriptionID)
	prepareClient(&client.Client, authorizer, partnerID)

	return &client, nil
}

func newGroupsClient(authorizer autorest.Authorizer, subscriptionID, partnerID string) (*resources.GroupsClient, error) {
	client := resources.NewGroupsClient(subscriptionID)
	prepareClient(&client.Client, authorizer, partnerID)

	return &client, nil
}

func newUsageClient(authorizer autorest.Authorizer, subscriptionID, partnerID string) (*compute.UsageClient, error) {
	client := compute.NewUsageClient(subscriptionID)
	prepareClient(&client.Client, authorizer, partnerID)

	return &client, nil
}

func newVirtualNetworkGatewayConnectionsClient(authorizer autorest.Authorizer, subscriptionID, partnerID string) (*network.VirtualNetworkGatewayConnectionsClient, error) {
	client := network.NewVirtualNetworkGatewayConnectionsClient(subscriptionID)
	prepareClient(&client.Client, authorizer, partnerID)

	return &client, nil
}

func newVirtualMachineScaleSetVMsClient(authorizer autorest.Authorizer, subscriptionID, partnerID string) (*compute.VirtualMachineScaleSetVMsClient, error) {
	client := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID)
	prepareClient(&client.Client, authorizer, partnerID)

	return &client, nil
}
func newApplicationsClient(clientID, clientSecret, tenantID, partnerID string) (*graphrbac.ApplicationsClient, error) {
	credentials := auth.ClientCredentialsConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
		Resource:     azure.PublicCloud.GraphEndpoint, // This Endpoint is different than using regular ClientCredentialsConfig
		AADEndpoint:  azure.PublicCloud.ActiveDirectoryEndpoint,
	}
	authorizer, err := credentials.Authorizer()
	if err != nil {
		return &graphrbac.ApplicationsClient{}, microerror.Mask(err)
	}

	client := graphrbac.NewApplicationsClient(tenantID)
	prepareClient(&client.Client, authorizer, partnerID)

	return &client, nil
}

func removeElementFromSlice(xs []int, x int) []int {
	for i, v := range xs {
		if v == x {
			// Shift end of slice to the left by one.
			copy(xs[i:], xs[i+1:])
			// Truncate the last element.
			xs = xs[:len(xs)-1]
			// Call it a day.
			break
		}
	}

	return xs
}
