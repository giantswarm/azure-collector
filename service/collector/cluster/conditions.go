package cluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v2/pkg/label"
	aeconditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Conditions struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

var (
	clusterStatus = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "cluster", "status"),
		"Latest cluster status conditions as provided by the Cluster CR status.",
		[]string{
			"cluster_id",
			"release_version",
			"status",
		},
		nil,
	)
)

func NewConditions(ctrlClient client.Client, logger micrologger.Logger) (*Conditions, error) {
	if ctrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "ctrlClient must not be empty")
	}
	if logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	c := &Conditions{
		ctrlClient: ctrlClient,
		logger:     logger,
	}

	return c, nil
}

func (c *Conditions) Collect(ctx context.Context, cluster *capiv1alpha3.Cluster, ch chan<- prometheus.Metric) error {
	c.logger.Debugf(ctx, "collecting cluster condition status metric for cluster %q", cluster.Name)

	releaseVersion, ok := cluster.Labels[label.ReleaseVersion]
	if !ok {
		c.logger.Debugf(ctx, "Cluster %#q has no %#q label. Skipping", cluster.Name, label.ReleaseVersion)
		return nil
	}

	var isCreating float64
	if conditions.IsTrue(cluster, aeconditions.CreatingCondition) {
		isCreating = 1
	}
	ch <- prometheus.MustNewConstMetric(
		clusterStatus,
		prometheus.GaugeValue,
		isCreating,
		cluster.Name,
		releaseVersion,
		string(aeconditions.CreatingCondition),
	)

	var isUpgrading float64
	if conditions.IsTrue(cluster, aeconditions.UpgradingCondition) {
		isUpgrading = 1
	}
	ch <- prometheus.MustNewConstMetric(
		clusterStatus,
		prometheus.GaugeValue,
		isUpgrading,
		cluster.Name,
		releaseVersion,
		string(aeconditions.UpgradingCondition),
	)

	c.logger.Debugf(ctx, "collected cluster condition status metric for cluster %q", cluster.Name)

	return nil
}

func (c *Conditions) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterStatus
	return nil
}
