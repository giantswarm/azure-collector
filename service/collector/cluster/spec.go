package cluster

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

type ClusterCollector interface {
	Collect(ctx context.Context, cr *capiv1alpha3.Cluster, ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc) error
}
