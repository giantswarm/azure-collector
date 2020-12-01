package collector

import (
	"context"

	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/azure-collector/v2/service/credential"
)

var (
	usageCurrentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "usage", "current"),
		"Current usage of specific Quotas as defined by Azure.",
		[]string{
			"name",
			"subscription",
		},
		nil,
	)
	usageLimitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "usage", "limit"),
		"Usage limit of specific Quotas as defined by Azure.",
		[]string{
			"name",
			"subscription",
		},
		nil,
	)
	scrapeErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{Namespace: MetricsNamespace, Subsystem: "usage", Name: "scrape_error",
			Help: "Total number of times compute resource usage information scraping returned an error.",
		})
)

type UsageConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	Location   string
	GSTenantID string
}

type Usage struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	usageScrapeError prometheus.Counter

	location   string
	gsTenantID string
}

func init() {
	prometheus.MustRegister(scrapeErrorCounter)
}

// NewUsage exposes metrics about the quota usage on Azure so we can alert when we are reaching the quota limits.
// It exposes quota metrics for every subscription found in the "credential-*" secrets of the control plane.
func NewUsage(config UsageConfig) (*Usage, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	u := &Usage{
		g8sClient:        config.G8sClient,
		k8sClient:        config.K8sClient,
		logger:           config.Logger,
		usageScrapeError: scrapeErrorCounter,
		location:         config.Location,
		gsTenantID:       config.GSTenantID,
	}

	return u, nil
}

func (u *Usage) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	clientSets, err := credential.GetAzureClientSetsFromCredentialSecretsBySubscription(ctx, u.k8sClient, u.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	// We track usage metrics for each client labeled by subscription.
	// That way we prevent duplicated metrics.
	for subscriptionID, azureClientSet := range clientSets {
		r, err := azureClientSet.UsageClient.List(ctx, u.location)
		if err != nil {
			u.logger.Errorf(ctx, err, "an error occurred during the scraping of current compute resource usage information")
			u.usageScrapeError.Inc()
		} else {
			for r.NotDone() {
				for _, v := range r.Values() {
					ch <- prometheus.MustNewConstMetric(
						usageCurrentDesc,
						prometheus.GaugeValue,
						float64(*v.CurrentValue),
						*v.Name.LocalizedValue,
						subscriptionID,
					)
					ch <- prometheus.MustNewConstMetric(
						usageLimitDesc,
						prometheus.GaugeValue,
						float64(*v.Limit),
						*v.Name.LocalizedValue,
						subscriptionID,
					)
				}

				err := r.NextWithContext(ctx)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	return nil
}

func (u *Usage) Describe(ch chan<- *prometheus.Desc) error {
	ch <- usageCurrentDesc
	ch <- usageLimitDesc
	return nil
}
