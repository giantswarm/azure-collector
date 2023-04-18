package cluster

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	expcapiv1beta1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v3/internal/capzcredentials"
)

type NodePools struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

var (
	clusterNodePools = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "cluster", "node_pools"),
		"Exposes the number of node pools in a cluster",
		[]string{
			"cluster_id",
		},
		nil,
	)

	clusterWorkers = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "cluster", "worker_nodes"),
		"Exposes the number of worker nodes in a cluster",
		[]string{
			"cluster_id",
		},
		nil,
	)
)

func NewNodePools(ctrlClient client.Client, logger micrologger.Logger) (*NodePools, error) {
	if ctrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "ctrlClient must not be empty")
	}
	if logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	c := &NodePools{
		ctrlClient: ctrlClient,
		logger:     logger,
	}

	return c, nil
}

func (n *NodePools) Collect(ctx context.Context, cluster *capiv1beta1.Cluster, ch chan<- prometheus.Metric) error {
	var nodePoolsCount int
	var currentWorkersCount int64
	{
		nps := expcapiv1beta1.MachinePoolList{}
		err := n.ctrlClient.List(ctx, &nps, client.MatchingLabels{label.Cluster: cluster.Name})
		if err != nil {
			return microerror.Mask(err)
		}

		nodePoolsCount = len(nps.Items)

		for _, np := range nps.Items {
			// Get VMSS regarding to this NP and get current size.
			azureCredentials, err := capzcredentials.GetAzureCredentialsFromMetadata(ctx, n.ctrlClient, cluster.ObjectMeta)
			if err != nil {
				n.logger.Errorf(ctx, err, "Unable to get azure credentials for cluster %q", cluster.Name)
				continue
			}

			var vmssClient compute.VirtualMachineScaleSetsClient
			{
				settings := auth.NewClientCredentialsConfig(azureCredentials.ClientID, azureCredentials.ClientSecret, azureCredentials.TenantID)
				authorizer, err := settings.Authorizer()
				if err != nil {
					n.logger.Errorf(ctx, err, "Unable to use azure credentials for cluster %q", cluster.Name)
					continue
				}
				vmssClient = compute.NewVirtualMachineScaleSetsClient(azureCredentials.SubscriptionID)
				vmssClient.Client.Authorizer = authorizer
			}

			resp, err := vmssClient.Get(ctx, cluster.Name, fmt.Sprintf("nodepool-%s", np.Name))
			if err != nil {
				n.logger.Errorf(ctx, err, "Unable to get vmss for np %q in cluster %q", np.Name, cluster.Name)
				continue
			}

			if resp.Sku != nil && resp.Sku.Capacity != nil {
				currentWorkersCount += *resp.Sku.Capacity
			}
		}
	}

	ch <- prometheus.MustNewConstMetric(
		clusterNodePools,
		prometheus.GaugeValue,
		float64(nodePoolsCount),
		cluster.Name,
	)

	ch <- prometheus.MustNewConstMetric(
		clusterWorkers,
		prometheus.GaugeValue,
		float64(currentWorkersCount),
		cluster.Name,
	)

	return nil
}

func (n *NodePools) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterNodePools
	ch <- clusterWorkers
	return nil
}
