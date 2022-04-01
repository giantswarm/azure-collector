package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/azure-collector/v2/client"
	"github.com/giantswarm/azure-collector/v2/service/credential"
)

const (
	labelClientId        = "client_id"
	labelSubscriptionId  = "subscription_id"
	labelTenantId        = "tenant_id"
	labelApplicationId   = "application_id"
	labelApplicationName = "application_name"
	labelFailureReason   = "reason"
	labelSecretKeyID     = "secret_key_id"
)

var (
	spExpirationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "service_principal_token", "expiration"),
		"Expiration date for Azure Access Tokens.",
		[]string{
			labelClientId,
			labelSubscriptionId,
			labelTenantId,
			labelApplicationId,
			labelApplicationName,
			labelSecretKeyID,
		},
		nil,
	)

	spExpirationFailedScrapeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "service_principal_token", "check_failed"),
		"Unable to retrieve informations about the service principal expiration date.",
		[]string{
			labelClientId,
			labelSubscriptionId,
			labelTenantId,
			labelFailureReason,
		},
		nil,
	)
)

type SPExpirationConfig struct {
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
	GSTenantID string
}

type SPExpiration struct {
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
	gsTenantID string
}

// NewSPExpiration exposes metrics about the expiration date of Azure Service Principals.
// It exposes metrcis about the Service Principals found in the "credential-*" secrets of the control plane.
func NewSPExpiration(config SPExpirationConfig) (*SPExpiration, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	v := &SPExpiration{
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return v, nil
}

func (v *SPExpiration) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	azureClientSets, err := credential.GetAzureClientSetsFromCredentialSecrets(ctx, v.k8sClient, v.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(azureClientSets) < 1 {
		v.logger.Debugf(ctx, "No clusters, skipping SP expiration collector")
		return nil
	}

	type scrapeErrorWithReason struct {
		AzureClientSetConfig *client.AzureClientSetConfig
		Reason               string
	}

	failedScrapes := make(map[string]scrapeErrorWithReason)

	// Use one arbitrary client set (we don't care which one) and use it to list all service principals on the GiantSwarm Active Directory.
	for azureClientSetConfig, clientSet := range azureClientSets {
		apps, err := clientSet.ApplicationsClient.ListComplete(ctx, "")
		if err != nil {
			reason := "unknown"
			// Catch if error message contains 'is expired' as that means the credentials are, well, expired.
			if IsCredentialsExpiredError(err) {
				reason = "expired"
			} else if IsForbiddenError(err) {
				reason = "forbidden"
			}

			// Ignore but log
			v.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unable to list applications using client %#q", azureClientSetConfig.ClientID), "stack", microerror.JSON(err), "gsTenantID", v.gsTenantID)
			failedScrapes[azureClientSetConfig.ClientID] = scrapeErrorWithReason{
				AzureClientSetConfig: azureClientSetConfig,
				Reason:               reason,
			}
			continue
		}

		for apps.NotDone() {
			app := apps.Value()
			for _, pc := range *app.PasswordCredentials {
				ch <- prometheus.MustNewConstMetric(
					spExpirationDesc,
					prometheus.GaugeValue,
					float64(pc.EndDate.Unix()),
					azureClientSetConfig.ClientID,
					azureClientSetConfig.SubscriptionID,
					azureClientSetConfig.TenantID,
					*app.AppID,
					*app.DisplayName,
					*pc.KeyID,
				)
			}

			if err := apps.NextWithContext(ctx); err != nil {
				return microerror.Mask(err)
			}
		}

		// We just need to list service principals once, so we can leave the loop.
		break
	}

	// Send metrics for failed scrapes as well
	for _, failedScrape := range failedScrapes {
		ch <- prometheus.MustNewConstMetric(
			spExpirationFailedScrapeDesc,
			prometheus.GaugeValue,
			float64(1),
			failedScrape.AzureClientSetConfig.ClientID,
			failedScrape.AzureClientSetConfig.SubscriptionID,
			failedScrape.AzureClientSetConfig.TenantID,
			failedScrape.Reason,
		)
	}

	return nil
}

func (v *SPExpiration) Describe(ch chan<- *prometheus.Desc) error {
	ch <- spExpirationDesc
	ch <- spExpirationFailedScrapeDesc
	return nil
}
