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

	gsTenantID = "31f75bf9-3d8c-4691-95c0-83dd71613db8"
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
	var collectors []collector.Interface

	{
		clusterCollectors, err := cluster.NewCollectors(config.K8sClient.CtrlClient(), config.Logger)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		conditions, err := cluster.NewConditions(config.K8sClient.CtrlClient(), config.Logger)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		nodepools, err := cluster.NewNodePools(config.K8sClient.CtrlClient(), config.Logger)
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
		clusterCollectors.Add(nodepools)
		clusterCollectors.Add(releases)
		clusterCollectors.Add(transition)
		collectors = append(collectors, clusterCollectors)
	}

	{
		c := DeploymentConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		deploymentCollector, err := NewDeployment(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, deploymentCollector)
	}

	{
		c := LoadBalancerConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		loadBalancerCollector, err := NewLoadBalancer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, loadBalancerCollector)
	}

	{
		c := ResourceGroupConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		resourceGroupCollector, err := NewResourceGroup(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, resourceGroupCollector)
	}

	{
		c := UsageConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
			Location:   config.Location,
			GSTenantID: config.GSTenantID,
		}

		usageCollector, err := NewUsage(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, usageCollector)
	}

	{
		c := RateLimitConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Location:   config.Location,
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		rateLimitCollector, err := NewRateLimit(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, rateLimitCollector)
	}

	{
		if config.GSTenantID == gsTenantID {
			c := SPExpirationConfig{
				CtrlClient: config.K8sClient.CtrlClient(),
				Logger:     config.Logger,
				GSTenantID: config.GSTenantID,
			}

			spExpirationCollector, err := NewSPExpiration(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			collectors = append(collectors, spExpirationCollector)
		}
	}

	{
		c := VMSSRateLimitConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
			GSTenantID: config.GSTenantID,
		}

		vmssRateLimitCollector, err := NewVMSSRateLimit(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, vmssRateLimitCollector)
	}

	{
		c := VPNConnectionConfig{
			CtrlClient:       config.K8sClient.CtrlClient(),
			InstallationName: config.ControlPlaneResourceGroup,
			Logger:           config.Logger,
			GSTenantID:       config.GSTenantID,
		}

		vpnConnectionCollector, err := NewVPNConnection(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		collectors = append(collectors, vpnConnectionCollector)
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: collectors,
			Logger:     config.Logger,
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
