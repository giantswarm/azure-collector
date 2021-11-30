package cluster

import (
	"context"
	"strconv"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			"has_node_pools",
		},
		nil,
	)

	clusterWorkers = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "cluster", "workers"),
		"Exposes the number of worker nodes in a cluster",
		[]string{
			"cluster_id",
			"has_workers",
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

func (n *NodePools) Collect(ctx context.Context, cluster *capiv1alpha3.Cluster, ch chan<- prometheus.Metric) error {
	nodePoolsCount := 1
	currentWorkersCount := 4
	ch <- prometheus.MustNewConstMetric(
		clusterNodePools,
		prometheus.GaugeValue,
		float64(nodePoolsCount),
		cluster.Name,
		strconv.FormatBool(nodePoolsCount > 0),
	)

	ch <- prometheus.MustNewConstMetric(
		clusterWorkers,
		prometheus.GaugeValue,
		float64(currentWorkersCount),
		cluster.Name,
		strconv.FormatBool(currentWorkersCount > 0),
	)

	return nil
}

func (n *NodePools) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterNodePools
	return nil
}
