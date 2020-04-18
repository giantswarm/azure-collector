package collector

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/azure-collector/client"
	"github.com/giantswarm/azure-collector/service/credential"
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
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type SPExpiration struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func NewSPExpiration(config SPExpirationConfig) (*SPExpiration, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &SPExpiration{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return v, nil
}

func (v *SPExpiration) Collect(ch chan<- prometheus.Metric) error {
	azureClientSets, err := credential.GetAzureClientSetsFromCredentialSecrets(v.k8sClient)
	if err != nil {
		return microerror.Mask(err)
	}

	failedScrapes := make(map[string]*client.AzureClientSetConfig)

	for azureClientSetConfig := range azureClientSets {
		ctx := context.Background()

		c, err := v.getApplicationsClient(azureClientSetConfig)
		if err != nil {
			// Ignore but log
			v.logger.LogCtx(ctx, "level", "warning", "message", "Unable to create an applications client: ", err.Error())
			failedScrapes[azureClientSetConfig.ClientID] = azureClientSetConfig
			continue
		}

		apps, err := c.ListComplete(ctx, fmt.Sprintf("appId eq '%s'", azureClientSetConfig.ClientID))
		if err != nil {
			// Ignore but log
			v.logger.LogCtx(ctx, "level", "warning", "message", "Unable to get application: ", err.Error())
			failedScrapes[azureClientSetConfig.ClientID] = azureClientSetConfig
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

func (v *SPExpiration) getApplicationsClient(azureClientSetConfig *client.AzureClientSetConfig) (*graphrbac.ApplicationsClient, error) {
	c := graphrbac.NewApplicationsClient(azureClientSetConfig.TenantID)
	a, err := v.getGraphAuthorizer(azureClientSetConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c.Authorizer = a

	return &c, nil
}

func (v *SPExpiration) getAuthorizerForResource(azureClientSetConfig *client.AzureClientSetConfig, resource string) (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error

	env, err := azure.EnvironmentFromName(azureClientSetConfig.EnvironmentName)
	if err != nil {
		return a, microerror.Mask(err)
	}

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, azureClientSetConfig.TenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig, azureClientSetConfig.ClientID, azureClientSetConfig.ClientSecret, resource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	a = autorest.NewBearerAuthorizer(token)

	return a, err
}

func (v *SPExpiration) getGraphAuthorizer(azureClientSetConfig *client.AzureClientSetConfig) (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error

	env, err := azure.EnvironmentFromName(azureClientSetConfig.EnvironmentName)
	if err != nil {
		return a, microerror.Mask(err)
	}

	a, err = v.getAuthorizerForResource(azureClientSetConfig, env.GraphEndpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return a, err
}
