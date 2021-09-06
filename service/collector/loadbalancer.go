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
	G8sClient  versioned.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
	GSTenantID string
}

type LoadBalancer struct {
	g8sClient  versioned.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
	gsTenantID string
}

// NewLoadBalancer exposes metrics about the 'kubernetes' load balancer used to Kubernetes services with type LoadBalancer.
func NewLoadBalancer(config LoadBalancerConfig) (*LoadBalancer, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	d := &LoadBalancer{
		g8sClient:  config.G8sClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return d, nil
}

func (d *LoadBalancer) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	azureClientSets, err := credential.GetAzureClientSetsByCluster(ctx, d.k8sClient, d.g8sClient, d.gsTenantID)
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
