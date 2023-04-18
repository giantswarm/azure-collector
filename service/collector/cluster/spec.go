package cluster

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

type ClusterCollector interface {
	Collect(ctx context.Context, cr *capiv1beta1.Cluster, ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc) error
}
