package cluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MetricsNamespace = "azure_operator"
)

type Collectors struct {
	ctrlClient client.Client
	logger     micrologger.Logger

	collectors []ClusterCollector
}

func NewCollectors(ctrlClient client.Client, logger micrologger.Logger) (*Collectors, error) {
	if ctrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "ctrlClient must not be empty")
	}
	if logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	c := &Collectors{
		ctrlClient: ctrlClient,
		logger:     logger,
	}

	return c, nil
}

func (c *Collectors) Add(cc ClusterCollector) {
	c.collectors = append(c.collectors, cc)
}

func (c *Collectors) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	clusters := &capiv1beta1.ClusterList{}
	{
		err := c.ctrlClient.List(ctx, clusters, client.InNamespace(metav1.NamespaceAll))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, cr := range clusters.Items {
		for _, collector := range c.collectors {
			err := collector.Collect(ctx, &cr, ch) //nolint:gosec
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (c *Collectors) Describe(ch chan<- *prometheus.Desc) error {
	for _, collector := range c.collectors {
		err := collector.Describe(ch)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
