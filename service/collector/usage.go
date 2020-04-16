package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/azure-collector/client"
	"github.com/giantswarm/azure-collector/service/credential"
)

var (
	usageCurrentDesc = prometheus.NewDesc(
		prometheus.BuildFQName("azure_operator", "usage", "current"),
		"Current usage of specific Quotas as defined by Azure.",
		[]string{
			"name",
			"subscription",
		},
		nil,
	)
	usageLimitDesc = prometheus.NewDesc(
		prometheus.BuildFQName("azure_operator", "usage", "limit"),
		"Usage limit of specific Quotas as defined by Azure.",
		[]string{
			"name",
			"subscription",
		},
		nil,
	)
	scrapeErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{Namespace: "azure_operator", Subsystem: "usage", Name: "scrape_error",
			Help: "Total number of times compute resource usage information scraping returned an error.",
		})
)

type UsageConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	Location               string
	CPAzureClientSetConfig client.AzureClientSetConfig
}

type Usage struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	usageScrapeError prometheus.Counter

	location               string
	cpAzureClientSetConfig client.AzureClientSetConfig
}

func init() {
	prometheus.MustRegister(scrapeErrorCounter)
}

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

	u := &Usage{
		g8sClient:              config.G8sClient,
		k8sClient:              config.K8sClient,
		logger:                 config.Logger,
		usageScrapeError:       scrapeErrorCounter,
		location:               config.Location,
		cpAzureClientSetConfig: config.CPAzureClientSetConfig,
	}

	return u, nil
}

func (u *Usage) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	clientSets, err := credential.GetAzureClientSetsFromCredentialSecretsBySubscription(u.k8sClient, u.cpAzureClientSetConfig.EnvironmentName)
	if err != nil {
		return microerror.Mask(err)
	}

	// The operator potentially uses a different set of credentials than
	// tenant clusters, so we add the operator credentials as well.
	operatorClientSet, err := client.NewAzureClientSet(u.cpAzureClientSetConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	clientSets[u.cpAzureClientSetConfig.SubscriptionID] = operatorClientSet

	// We track usage metrics for each client labeled by subscription.
	// That way we prevent duplicated metrics.
	for subscriptionID, azureClientSet := range clientSets {
		r, err := azureClientSet.UsageClient.List(ctx, u.location)
		if err != nil {
			u.logger.Log("level", "warning", "message", "an error occurred during the scraping of current compute resource usage information", "stack", fmt.Sprintf("%v", err))
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
