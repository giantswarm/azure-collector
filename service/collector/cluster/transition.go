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

type TransitionTime struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

var (
	clusterTransitionCreateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "cluster", "create_transition"),
		"Latest cluster creation transition.",
		[]string{
			"cluster_id",
			"release_version",
		},
		nil,
	)
)

func NewTransitionTime(ctrlClient client.Client, logger micrologger.Logger) (*TransitionTime, error) {
	if ctrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "ctrlClient must not be empty")
	}
	if logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	u := &TransitionTime{
		ctrlClient: ctrlClient,
		logger:     logger,
	}

	return u, nil
}

func (t *TransitionTime) Collect(ctx context.Context, cluster *capiv1alpha3.Cluster, ch chan<- prometheus.Metric) error {
	t.logger.Debugf(ctx, "collected cluster transition time metric for cluster %q", cluster.Name)

	releaseVersion, ok := cluster.Labels[label.ReleaseVersion]
	if !ok {
		t.logger.Debugf(ctx, "Cluster %#q has no %#q label. Skipping", cluster.Name, label.ReleaseVersion)
		return nil
	}

	if !conditions.IsFalse(cluster, aeconditions.CreatingCondition) {
		t.logger.Debugf(ctx, "Cluster %#q has no %#q condition or it's still being created. Skipping", cluster.Name, aeconditions.CreatingCondition)
		return nil
	}

	creatingLastTransition := conditions.GetLastTransitionTime(cluster, aeconditions.CreatingCondition)
	ch <- prometheus.MustNewConstMetric(
		clusterTransitionCreateDesc,
		prometheus.GaugeValue,
		creatingLastTransition.Sub(cluster.CreationTimestamp.Time).Seconds(),
		cluster.Name,
		releaseVersion,
	)

	t.logger.Debugf(ctx, "collected cluster transition time metric for cluster %q", cluster.Name)
	return nil
}

func (t *TransitionTime) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	return nil
}
