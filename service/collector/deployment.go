package collector

import (
	"context"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v2/service/credential"
)

const (
	statusCanceled  = "Canceled"
	statusFailed    = "Failed"
	statusRunning   = "Running"
	statusSucceeded = "Succeeded"
)

var (
	deploymentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "deployment", "status"),
		"Cluster status condition as provided by the CR status.",
		[]string{
			"cluster_id",
			"deployment_name",
			"status",
		},
		nil,
	)
)

type DeploymentConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
	GSTenantID string
}

type Deployment struct {
	ctrlClient client.Client
	logger     micrologger.Logger
	gsTenantID string
}

// NewDeployment exposes metrics about the Azure ARM Deployments for every cluster on this installation.
// It finds the cluster in the control plane, and uses the cluster Azure credentials to find the Deployments info.
func NewDeployment(config DeploymentConfig) (*Deployment, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	d := &Deployment{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return d, nil
}

func (d *Deployment) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	azureClientSets, err := credential.GetAzureClientSetsByCluster(ctx, d.ctrlClient, d.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	for clusterID, azureClientSet := range azureClientSets {
		r, err := azureClientSet.DeploymentsClient.ListByResourceGroup(context.Background(), clusterID, "", to.Int32Ptr(100))
		if err != nil {
			return microerror.Mask(err)
		}

		for r.NotDone() {
			for _, v := range r.Values() {
				ch <- prometheus.MustNewConstMetric(
					deploymentDesc,
					prometheus.GaugeValue,
					float64(matchedStringToInt(statusCanceled, *v.Properties.ProvisioningState)),
					clusterID,
					*v.Name,
					statusCanceled,
				)
				ch <- prometheus.MustNewConstMetric(
					deploymentDesc,
					prometheus.GaugeValue,
					float64(matchedStringToInt(statusFailed, *v.Properties.ProvisioningState)),
					clusterID,
					*v.Name,
					statusFailed,
				)
				ch <- prometheus.MustNewConstMetric(
					deploymentDesc,
					prometheus.GaugeValue,
					float64(matchedStringToInt(statusRunning, *v.Properties.ProvisioningState)),
					clusterID,
					*v.Name,
					statusRunning,
				)
				ch <- prometheus.MustNewConstMetric(
					deploymentDesc,
					prometheus.GaugeValue,
					float64(matchedStringToInt(statusSucceeded, *v.Properties.ProvisioningState)),
					clusterID,
					*v.Name,
					statusSucceeded,
				)
			}

			err := r.NextWithContext(ctx)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (d *Deployment) Describe(ch chan<- *prometheus.Desc) error {
	ch <- deploymentDesc
	return nil
}

func matchedStringToInt(a, b string) int {
	if a == b {
		return 1
	}

	return 0
}
