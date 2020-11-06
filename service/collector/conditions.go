package collector

import (
	"context"
	"fmt"

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

type ConditionsConfig struct {
	CtrlClient ctrlclient.Client
	Logger     micrologger.Logger
}

type Conditions struct {
	ctrlClient ctrlclient.Client
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

func NewConditions(config ConditionsConfig) (*Conditions, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	u := &Conditions{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return u, nil
}

func (u *Conditions) Collect(ch chan<- prometheus.Metric) error {
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
			u.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Cluster %#q has no %#q label. Skipping", cluster.Name, label.ReleaseVersion))
			continue
		}

		var isCreating float64
		if conditions.IsTrue(&cluster, aeconditions.CreatingCondition) {
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
		if conditions.IsTrue(&cluster, aeconditions.UpgradingCondition) {
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
	}

	return nil
}

func (u *Conditions) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterStatus
	return nil
}
