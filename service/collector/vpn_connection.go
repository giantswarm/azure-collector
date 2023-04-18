package collector

import (
	"context"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v3/service/credential"
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
	CtrlClient       client.Client
	InstallationName string
	Logger           micrologger.Logger
	GSTenantID       string
}

type VPNConnection struct {
	ctrlClient       client.Client
	installationName string
	logger           micrologger.Logger
	gsTenantID       string
}

func NewVPNConnection(config VPNConnectionConfig) (*VPNConnection, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	v := &VPNConnection{
		ctrlClient:       config.CtrlClient,
		installationName: config.InstallationName,
		logger:           config.Logger,
		gsTenantID:       config.GSTenantID,
	}

	return v, nil
}

func (v *VPNConnection) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	azureClientSets, err := credential.GetAzureClientSetsByCluster(ctx, v.ctrlClient, v.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	for clusterID, azureClientSet := range azureClientSets {
		connections, err := azureClientSet.VirtualNetworkGatewayConnectionsClient.ListComplete(ctx, clusterID)
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
				connection, err := azureClientSet.VirtualNetworkGatewayConnectionsClient.Get(ctx, clusterID, connectionName)
				if err != nil {
					return microerror.Mask(err)
				}

				// We ignore customer's VPN gateways by filtering the VPN gateway name.
				// We use the installation name as the VPN gateway name.
				if to.String(connection.ID) != v.installationName {
					return nil
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
	}

	return nil
}

func (v *VPNConnection) Describe(ch chan<- *prometheus.Desc) error {
	ch <- vpnConnectionDesc
	return nil
}
