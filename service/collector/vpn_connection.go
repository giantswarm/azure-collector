package collector

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/azure-collector/v2/client"
	"github.com/giantswarm/azure-collector/v2/service/credential"
)

var (
	vpnConnectionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "vpn_connection", "info"),
		"VPN connection informations.",
		[]string{
			"id",
			"name",
			"location",
			"connection_type",
			"connection_status",
			"provisioning_state",
		},
		nil,
	)
)

type VPNConnectionConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ResourceGroup string
	GSTenantID    string
}

type VPNConnection struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	resourceGroup string
	gsTenantID    string
}

func NewVPNConnection(config VPNConnectionConfig) (*VPNConnection, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ResourceGroup == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ResourceGroup must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	v := &VPNConnection{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		resourceGroup: config.ResourceGroup,
		gsTenantID:    config.GSTenantID,
	}

	return v, nil
}

func (v *VPNConnection) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	vpnConnectionClient, err := v.getVPNConnectionsClient(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	resourceGroup := v.resourceGroup
	connections, err := vpnConnectionClient.ListComplete(ctx, resourceGroup)
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for connections.NotDone() {
		c := connections.Value()
		connectionName := to.String(c.Name)

		// ConnectionStatus returned by the API when listing connections is always empty.
		// Details for each connection must be requested in order to get a value for ConnectionStatus.
		g.Go(func() error {
			connection, err := vpnConnectionClient.Get(ctx, resourceGroup, connectionName)
			if err != nil {
				return microerror.Mask(err)
			}

			ch <- prometheus.MustNewConstMetric(
				vpnConnectionDesc,
				prometheus.GaugeValue,
				1,
				to.String(connection.ID),
				connectionName,
				to.String(connection.Location),
				string(connection.ConnectionType),
				string(connection.ConnectionStatus),
				string(connection.ProvisioningState),
			)

			return nil
		})

		if err := connections.NextWithContext(ctx); err != nil {
			return microerror.Mask(err)
		}
	}

	if err := g.Wait(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (v *VPNConnection) Describe(ch chan<- *prometheus.Desc) error {
	ch <- vpnConnectionDesc
	return nil
}

func (v *VPNConnection) getVPNConnectionsClient(ctx context.Context) (*network.VirtualNetworkGatewayConnectionsClient, error) {
	config, err := credential.GetAzureConfigFromSecretName(ctx, v.k8sClient, credential.CredentialDefault, credential.CredentialNamespace, v.gsTenantID)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	azureClients, err := client.NewAzureClientSet(*config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return azureClients.VirtualNetworkGatewayConnectionsClient, nil
}
