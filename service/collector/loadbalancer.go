package collector

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v3/service/credential"
)

var (
	loadBalancerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "load_balancer", "backend_pool_instances_count"),
		"The number of instances behind a backend pool.",
		[]string{
			"cluster_id",
			"load_balancer_name",
			"backend_pool_name",
		},
		nil,
	)
)

type LoadBalancerConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
	GSTenantID string
}

type LoadBalancer struct {
	ctrlClient client.Client
	logger     micrologger.Logger
	gsTenantID string
}

// NewLoadBalancer exposes metrics about the 'kubernetes' load balancer used to Kubernetes services with type LoadBalancer.
func NewLoadBalancer(config LoadBalancerConfig) (*LoadBalancer, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	d := &LoadBalancer{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return d, nil
}

func (d *LoadBalancer) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	azureClientSets, err := credential.GetAzureClientSetsByCluster(ctx, d.ctrlClient, d.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	lbNames := []string{"kubernetes", "kubernetes-internal"}

	for clusterID, azureClientSet := range azureClientSets {
		for _, lbName := range lbNames {
			lb, err := azureClientSet.LoadBalancersClient.Get(context.Background(), clusterID, lbName, "")
			if IsNotFound(err) {
				// Load balancer might be missing, all good.
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}

			if lb.BackendAddressPools != nil {
				for _, bp := range *lb.BackendAddressPools {
					if bp.BackendIPConfigurations != nil {
						ch <- prometheus.MustNewConstMetric(
							loadBalancerDesc,
							prometheus.GaugeValue,
							float64(len(*bp.BackendIPConfigurations)),
							clusterID,
							lbName,
							*bp.Name,
						)
					}
				}
			}
		}
	}

	return nil
}

func (d *LoadBalancer) Describe(ch chan<- *prometheus.Desc) error {
	ch <- loadBalancerDesc
	return nil
}
