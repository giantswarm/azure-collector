package collector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v3/pkg/project"
	"github.com/giantswarm/azure-collector/v3/service/credential"
)

const (
	remainingReadsHeaderName  = "x-ms-ratelimit-remaining-subscription-reads"
	remainingWritesHeaderName = "x-ms-ratelimit-remaining-subscription-writes"
	resourceGroupNamePrefix   = "azure-collector-empty-rg-for-metrics"
	metricsSubsystem          = "rate_limit"
)

var (
	readsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, metricsSubsystem, "reads"),
		"Remaining number of reads allowed.",
		[]string{
			"subscription",
			"clientid",
		},
		nil,
	)
	writesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, metricsSubsystem, "writes"),
		"Remaining number of writes allowed.",
		[]string{
			"subscription",
			"clientid",
		},
		nil,
	)
	readsErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: MetricsNamespace,
		Subsystem: metricsSubsystem,
		Name:      "reads_parsing_errors",
		Help:      "Errors trying to parse the remaining requests from the response header",
	})
	writesErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: MetricsNamespace,
		Subsystem: metricsSubsystem,
		Name:      "writes_parsing_errors",
		Help:      "Errors trying to parse the remaining requests from the response header",
	})
)

type RateLimitConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
	Location   string
	GSTenantID string
}

type RateLimit struct {
	ctrlClient client.Client
	logger     micrologger.Logger
	location   string
	gsTenantID string
}

func init() {
	prometheus.MustRegister(readsErrorCounter)
	prometheus.MustRegister(writesErrorCounter)
}

// NewRateLimit exposes metrics about the Azure resource group client rate limit.
// It creates and fetches a resource group. That way it can inspect the Azure API response to find rate limit headers.
// It uses the credentials found in the "credential-*" secrets of the control plane.
func NewRateLimit(config RateLimitConfig) (*RateLimit, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
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

	u := &RateLimit{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		location:   config.Location,
		gsTenantID: config.GSTenantID,
	}

	return u, nil
}

func (u *RateLimit) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	clientSets, err := credential.GetAzureClientSetsFromCredentialSecrets(ctx, u.ctrlClient, u.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	var doneSubscriptions []string

	// We track RateLimit metrics for each client labeled by SubscriptionID and
	// ClientID.
	// That way we prevent duplicated metrics.
	for clientConfig, clientSet := range clientSets {
		// We want to check only once per subscription
		if inArray(doneSubscriptions, clientSet.GroupsClient.SubscriptionID) {
			continue
		}

		// Remaining write requests can be retrieved sending a write request.
		var writes float64
		{
			resourceGroup := resources.Group{
				ManagedBy: to.StringPtr(project.Name()),
				Location:  to.StringPtr(u.location),
				Tags: map[string]*string{
					"collector": to.StringPtr(project.Name()),
				},
			}
			resourceGroup, err := clientSet.GroupsClient.CreateOrUpdate(ctx, u.getResourceGroupName(), resourceGroup)
			if err != nil {
				u.logger.Debugf(ctx, "clientid %#q gstenantid %#q tenantid %#q", clientConfig.ClientID, clientConfig.GSTenantID, clientConfig.TenantID)
				return microerror.Mask(err)
			}

			writes, err = strconv.ParseFloat(resourceGroup.Response.Header.Get(remainingWritesHeaderName), 64)
			if err != nil {
				u.logger.Errorf(ctx, err, "an error occurred parsing to float the value inside the rate limiting header for write requests")
				writes = 0
				writesErrorCounter.Inc()
			}

			ch <- prometheus.MustNewConstMetric(
				writesDesc,
				prometheus.GaugeValue,
				writes,
				clientSet.GroupsClient.SubscriptionID,
				clientConfig.ClientID,
			)

			doneSubscriptions = append(doneSubscriptions, clientSet.GroupsClient.SubscriptionID)
		}

		// Remaining read requests can be retrieved sending a read request.
		var reads float64
		{
			groupResponse, err := clientSet.GroupsClient.Get(ctx, u.getResourceGroupName())
			if err != nil {
				return microerror.Mask(err)
			}

			reads, err = strconv.ParseFloat(groupResponse.Response.Header.Get(remainingReadsHeaderName), 64)
			if err != nil {
				u.logger.Errorf(ctx, err, "an error occurred parsing to float the value inside the rate limiting header for read requests")
				reads = 0
				readsErrorCounter.Inc()
			}

			ch <- prometheus.MustNewConstMetric(
				readsDesc,
				prometheus.GaugeValue,
				reads,
				clientSet.GroupsClient.SubscriptionID,
				clientConfig.ClientID,
			)
		}
	}

	return nil
}

func (u *RateLimit) Describe(ch chan<- *prometheus.Desc) error {
	ch <- readsDesc
	ch <- writesDesc
	return nil
}

func (u *RateLimit) getResourceGroupName() string {
	return fmt.Sprintf("%s-%s", resourceGroupNamePrefix, u.location)
}
