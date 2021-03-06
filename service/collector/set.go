package collector

import (
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/azure-collector/v2/service/collector/cluster"
)

const (
	MetricsNamespace = "azure_operator"
)

type SetConfig struct {
	K8sClient                 k8sclient.Interface
	Location                  string
	Logger                    micrologger.Logger
	ControlPlaneResourceGroup string
	GSTenantID                string
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the iniitialization and prevents some weird import mess so we do not
// have to alias packages.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	var err error

	var clusterCollectors *cluster.Collectors
	{
		clusterCollectors, err = cluster.NewCollectors(config.K8sClient.CtrlClient(), config.Logger)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		conditions, err := cluster.NewConditions(config.K8sClient.CtrlClient(), config.Logger)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		releases, err := cluster.NewReleases(config.K8sClient.CtrlClient(), config.Logger)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		transition, err := cluster.NewTransitionTime(config.K8sClient.CtrlClient(), config.Logger)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterCollectors.Add(conditions)
		clusterCollectors.Add(releases)
		clusterCollectors.Add(transition)
	}

	var deploymentCollector *Deployment
	{
		c := DeploymentConfig{
			G8sClient:  config.K8sClient.G8sClient(),
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		deploymentCollector, err = NewDeployment(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceGroupCollector *ResourceGroup
	{
		c := ResourceGroupConfig{
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		resourceGroupCollector, err = NewResourceGroup(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var usageCollector *Usage
	{
		c := UsageConfig{
			G8sClient:  config.K8sClient.G8sClient(),
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
			Location:   config.Location,
			GSTenantID: config.GSTenantID,
		}

		usageCollector, err = NewUsage(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var rateLimitCollector *RateLimit
	{
		c := RateLimitConfig{
			G8sClient:  config.K8sClient.G8sClient(),
			K8sClient:  config.K8sClient.K8sClient(),
			Location:   config.Location,
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		rateLimitCollector, err = NewRateLimit(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var spExpirationCollector *SPExpiration
	{
		c := SPExpirationConfig{
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		spExpirationCollector, err = NewSPExpiration(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vmssRateLimitCollector *VMSSRateLimit
	{
		c := VMSSRateLimitConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		vmssRateLimitCollector, err = NewVMSSRateLimit(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vpnConnectionCollector *VPNConnection
	{
		c := VPNConnectionConfig{
			G8sClient:        config.K8sClient.G8sClient(),
			InstallationName: config.ControlPlaneResourceGroup,
			K8sClient:        config.K8sClient.K8sClient(),
			Logger:           config.Logger,
			GSTenantID:       config.GSTenantID,
		}

		vpnConnectionCollector, err = NewVPNConnection(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				clusterCollectors,
				deploymentCollector,
				resourceGroupCollector,
				rateLimitCollector,
				spExpirationCollector,
				usageCollector,
				vmssRateLimitCollector,
				vpnConnectionCollector,
			},
			Logger: config.Logger,
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Set{
		Set: collectorSet,
	}

	return s, nil
}
