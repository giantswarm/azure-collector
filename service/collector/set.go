package collector

import (
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
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

	var clusterTransitionTime *ClusterTransitionTime
	{
		c := ClusterTransitionTimeConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		clusterTransitionTime, err = NewClusterTransitionTime(c)
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
				clusterTransitionTime,
				deploymentCollector,
				resourceGroupCollector,
				usageCollector,
				rateLimitCollector,
				spExpirationCollector,
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
