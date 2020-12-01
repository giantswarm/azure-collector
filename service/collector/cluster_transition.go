package collector

import (
	"context"

	"github.com/giantswarm/apiextensions/v2/pkg/label"
	aeconditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterTransitionTimeConfig struct {
	CtrlClient ctrlclient.Client
	Logger     micrologger.Logger
}

type ClusterTransitionTime struct {
	ctrlClient ctrlclient.Client
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

func NewClusterTransitionTime(config ClusterTransitionTimeConfig) (*ClusterTransitionTime, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	u := &ClusterTransitionTime{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return u, nil
}

func (u *ClusterTransitionTime) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	clusters := &capiv1alpha3.ClusterList{}
	{
		err := u.ctrlClient.List(ctx, clusters, ctrlclient.InNamespace(metav1.NamespaceAll))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, cluster := range clusters.Items {
		releaseVersion, ok := cluster.Labels[label.ReleaseVersion]
		if !ok {
			u.logger.Debugf(ctx, "Cluster %#q has no %#q label. Skipping", cluster.Name, label.ReleaseVersion)
			continue
		}

		if !conditions.IsFalse(&cluster, aeconditions.CreatingCondition) {
			u.logger.Debugf(ctx, "Cluster %#q has no %#q condition or it's still being created. Skipping", cluster.Name, aeconditions.CreatingCondition)
			continue
		}

		creatingLastTransition := conditions.GetLastTransitionTime(&cluster, aeconditions.CreatingCondition)
		ch <- prometheus.MustNewConstMetric(
			clusterTransitionCreateDesc,
			prometheus.GaugeValue,
			creatingLastTransition.Sub(cluster.CreationTimestamp.Time).Seconds(),
			cluster.Name,
			releaseVersion,
		)
	}

	return nil
}

func (u *ClusterTransitionTime) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	return nil
}
