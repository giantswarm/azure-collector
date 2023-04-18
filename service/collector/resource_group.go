package collector

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v2/service/credential"
)

const (
	labelID        = "id"
	labelName      = "name"
	labelState     = "state"
	labelLocation  = "location"
	labelManagedBy = "managed_by"
)

var (
	resourceGroupDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, "resource_group", "info"),
		"Resource group information.",
		[]string{
			labelID,
			labelName,
			labelState,
			labelLocation,
			labelManagedBy,
		},
		nil,
	)

	gaugeValue float64 = 1
)

type ResourceGroupConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
	GSTenantID string
}

type ResourceGroup struct {
	ctrlClient client.Client
	logger     micrologger.Logger
	gsTenantID string
}

// NewResourceGroup exposes metrics on the existing resource groups for every subscription.
// It exposes metrcis about the subscriptions found in the "credential-*" secrets of the control plane.
func NewResourceGroup(config ResourceGroupConfig) (*ResourceGroup, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	r := &ResourceGroup{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return r, nil
}

func (r *ResourceGroup) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	clientSets, err := credential.GetAzureClientSetsFromCredentialSecretsBySubscription(ctx, r.ctrlClient, r.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range clientSets {
		clientSet := item

		g.Go(func() error {
			err := r.collectForClientSet(ctx, ch, clientSet.GroupsClient)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *ResourceGroup) collectForClientSet(ctx context.Context, ch chan<- prometheus.Metric, client *resources.GroupsClient) error {
	resultsPage, err := client.ListComplete(context.Background(), "", nil)
	if err != nil {
		return microerror.Mask(err)
	}

	for resultsPage.NotDone() {
		group := resultsPage.Value()
		ch <- prometheus.MustNewConstMetric(
			resourceGroupDesc,
			prometheus.GaugeValue,
			gaugeValue,
			to.String(group.ID),
			to.String(group.Name),
			getState(group),
			to.String(group.Location),
			to.String(group.ManagedBy),
		)

		if err := resultsPage.NextWithContext(ctx); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *ResourceGroup) Describe(ch chan<- *prometheus.Desc) error {
	ch <- resourceGroupDesc

	return nil
}

func getState(group resources.Group) string {
	if group.Properties != nil {
		return to.String(group.Properties.ProvisioningState)
	}

	return ""
}
