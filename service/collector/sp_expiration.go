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
		v.logger.LogCtx(ctx, "level", "debug", "message", "No clusters, skipping SP expiration collector")
		return nil
	}

	failedScrapes := make(map[string]*client.AzureClientSetConfig)

	for azureClientSetConfig, clientSet := range azureClientSets {
		spObjectId, err := clientSet.ApplicationsClient.GetServicePrincipalsIDByAppID(ctx, azureClientSetConfig.ClientID)
		if err != nil {
			// Ignore but log
			v.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unable to get client application ID for client %#q", azureClientSetConfig.ClientID), "stack", microerror.JSON(err), "gsTenantID", v.gsTenantID)
			failedScrapes[azureClientSetConfig.ClientID] = azureClientSetConfig
			continue
		}

		app, err := clientSet.ApplicationsClient.Get(ctx, *spObjectId.Value)
		if err != nil {
			// Ignore but log
			v.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unable to fetch details for service principal %#q", *spObjectId.Value), "stack", microerror.JSON(err), "gsTenantID", v.gsTenantID)
			failedScrapes[azureClientSetConfig.ClientID] = azureClientSetConfig
			continue
		}

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
	}

	// Send metrics for failed scrapes as well
	for _, azureClientSetConfig := range failedScrapes {
		ch <- prometheus.MustNewConstMetric(
			spExpirationFailedScrapeDesc,
			prometheus.GaugeValue,
			float64(1),
			azureClientSetConfig.ClientID,
			azureClientSetConfig.SubscriptionID,
			azureClientSetConfig.TenantID,
		)
	}

	return nil
}

func (v *SPExpiration) Describe(ch chan<- *prometheus.Desc) error {
	ch <- spExpirationDesc
	ch <- spExpirationFailedScrapeDesc
	return nil
}
