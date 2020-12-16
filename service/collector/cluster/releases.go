package cluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Releases struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

var (
	clusterRelease = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "cluster", "release"),
		"Cluster release version.",
		[]string{
			"cluster_id",
			"release_version",
		},
		nil,
	)
)

func NewReleases(ctrlClient client.Client, logger micrologger.Logger) (*Releases, error) {
	if ctrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "ctrlClient must not be empty")
	}
	if logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	c := &Releases{
		ctrlClient: ctrlClient,
		logger:     logger,
	}

	return c, nil
}

func (c *Releases) Collect(ctx context.Context, cluster *capiv1alpha3.Cluster, ch chan<- prometheus.Metric) error {
	releaseVersion, ok := cluster.Labels[label.ReleaseVersion]
	if !ok {
		c.logger.Debugf(ctx, "Cluster %#q has no %#q label. Skipping", cluster.Name, label.ReleaseVersion)
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		clusterRelease,
		prometheus.GaugeValue,
		1,
		cluster.Name,
		releaseVersion,
	)

	return nil
}

func (c *Releases) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterRelease
	return nil
}
