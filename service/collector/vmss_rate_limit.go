package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-collector/v2/client"
	"github.com/giantswarm/azure-collector/v2/service/collector/key"
	"github.com/giantswarm/azure-collector/v2/service/credential"
)

const (
	// Note that an API request can be subjected to multiple throttling policies.
	// There will be a separate x-ms-ratelimit-remaining-resource header for each policy.
	//
	// Here is a sample response to delete virtual machine scale set request.
	//
	// x-ms-ratelimit-remaining-resource: Microsoft.Compute/DeleteVMScaleSet3Min;107
	// x-ms-ratelimit-remaining-resource: Microsoft.Compute/DeleteVMScaleSet30Min;587
	// x-ms-ratelimit-remaining-resource: Microsoft.Compute/VMScaleSetBatchedVMRequests5Min;3704
	// x-ms-ratelimit-remaining-resource: Microsoft.Compute/VmssQueuedVMOperations;4720
	vmssVMListHeaderName = "X-Ms-Ratelimit-Remaining-Resource"
	vmssMetricsSubsystem = "rate_limit"

	credentialDefaultNamespace = "giantswarm"
	credentialDefaultName      = "credential-default"
)

var (
	vmssVMListDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, vmssMetricsSubsystem, "vmss_instance_list"),
		"Remaining number of VMSS VM list operations.",
		[]string{
			"subscription",
			"clientid",
			"countername",
		},
		nil,
	)
	vmssMeasuredCallsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(MetricsNamespace, vmssMetricsSubsystem, "vmss_measured"),
		"Number of calls we are making as returned by the Azure APIs during errorbody 429 incident.",
		[]string{
			"subscription",
			"clientid",
			"countername",
		},
		nil,
	)
	vmssVMListErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: MetricsNamespace,
		Subsystem: vmssMetricsSubsystem,
		Name:      "vmss_instance_list_parsing_errors",
		Help:      "Errors trying to parse the remaining requests from the response header",
	})
)

type VMSSRateLimitConfig struct {
	CtrlClient ctrlclient.Client
	Logger     micrologger.Logger
	GSTenantID string
}

type VMSSRateLimit struct {
	ctrlClient ctrlclient.Client
	logger     micrologger.Logger
	gsTenantID string
}

func init() {
	prometheus.MustRegister(vmssVMListErrorCounter)
}

func NewVMSSRateLimit(config VMSSRateLimitConfig) (*VMSSRateLimit, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	u := &VMSSRateLimit{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return u, nil
}

func (u *VMSSRateLimit) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Remove 429 from the retriable error codes.
	original := autorest.StatusCodesForRetry
	defer func() {
		autorest.StatusCodesForRetry = original
	}()
	var codes []int
	for code := range autorest.StatusCodesForRetry {
		if code != http.StatusTooManyRequests {
			codes = append(codes, code)
		}
	}
	autorest.StatusCodesForRetry = codes

	clustersSecret, err := u.getClusters(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var doneSubscriptions []string
	for cluster, secret := range clustersSecret {
		config, err := credential.GetAzureConfigFromSecret(secret, u.gsTenantID)
		if err != nil {
			return microerror.Mask(err)
		}

		// We want to check only once per subscription
		if inArray(doneSubscriptions, config.SubscriptionID) {
			u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Skipping Cluster %#q, its subscription was already collected", cluster))
			continue
		}

		azureClients, err := client.NewAzureClientSet(*config)
		if err != nil {
			return microerror.Mask(err)
		}

		result, err := azureClients.VirtualMachineScaleSetVMsClient.ListComplete(ctx, cluster, fmt.Sprintf("%s-master-%s", cluster, cluster), "", "", "")
		if IsThrottlingError(err) {
			u.collectMeasuredCallsFromResponse(ch, result, config.SubscriptionID, config.ClientID)
		} else if err != nil {
			u.logger.LogCtx(ctx, "level", "warning", "message", "Skipping", "clientid", config.ClientID, "subscriptionid", config.SubscriptionID, "tenantid", config.TenantID, "stack", microerror.JSON(err))
			continue
		}

		// Note that an API request can be subjected to multiple throttling policies.
		// There will be a separate x-ms-ratelimit-remaining-resource header for each policy.
		headers, ok := result.Response().Header[vmssVMListHeaderName]
		if !ok {
			u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Header %#q not found", vmssVMListHeaderName), "headers", result.Response().Header)
			u.logger.LogCtx(ctx, "level", "warning", "message", "Skipping", "clientid", config.ClientID, "subscriptionid", config.SubscriptionID, "tenantid", config.TenantID, "stack", microerror.JSON(err))
			vmssVMListErrorCounter.Inc()
			continue
		}

		// Example header value: "x-ms-ratelimit-remaining-resource: Microsoft.Compute/DeleteVMScaleSet3Min;107"
		for _, header := range headers {
			// Limits are a single comma separated string.
			tokens := strings.SplitN(header, ",", -1)
			for _, t := range tokens {
				// Each limit's name and value are separated by a semicolon.
				kv := strings.SplitN(t, ";", 2)
				if len(kv) != 2 {
					// We expect exactly two tokens, otherwise we consider this a parsing error.
					u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unexpected limit in header. Expected something like 'Microsoft.Compute/DeleteVMScaleSet3Min;107', got %#q", t))
					u.logger.LogCtx(ctx, "level", "warning", "message", "Skipping", "clientid", config.ClientID, "subscriptionid", config.SubscriptionID)
					vmssVMListErrorCounter.Inc()
					continue
				}

				// The second token must be a number or we don't know what we got from MS.
				val, err := strconv.ParseFloat(kv[1], 64)
				if err != nil {
					u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unexpected value in limit. Expected a number, got %v", kv[1]))
					u.logger.LogCtx(ctx, "level", "warning", "message", "Skipping", "clientid", config.ClientID, "subscriptionid", config.SubscriptionID)
					vmssVMListErrorCounter.Inc()
					continue
				}

				ch <- prometheus.MustNewConstMetric(
					vmssVMListDesc,
					prometheus.GaugeValue,
					val,
					config.SubscriptionID,
					config.ClientID,
					kv[0],
				)

				doneSubscriptions = append(doneSubscriptions, config.SubscriptionID)
			}
		}
	}

	return nil
}

func (u *VMSSRateLimit) getClusters(ctx context.Context) (map[string]*v1.Secret, error) {
	clustersSecret := make(map[string]*v1.Secret)
	azureConfigs := &providerv1alpha1.AzureConfigList{}
	{
		err := u.ctrlClient.List(ctx, azureConfigs, ctrlclient.InNamespace(metav1.NamespaceAll))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	for _, azureConfig := range azureConfigs.Items {
		secret := &v1.Secret{}
		err := u.ctrlClient.Get(ctx, ctrlclient.ObjectKey{Namespace: key.CredentialNamespace(azureConfig), Name: key.CredentialName(azureConfig)}, secret)
		if err != nil {
			u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Skipping AzureConfig %#q", azureConfig.Name), "stack", microerror.JSON(err))
			continue
		}
		clustersSecret[azureConfig.Name] = secret
	}

	clusters := &v1alpha3.ClusterList{}
	{
		err := u.ctrlClient.List(ctx, clusters, ctrlclient.InNamespace(metav1.NamespaceAll))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	for _, cluster := range clusters.Items {
		credentialSecret, err := u.getOrganizationCredentialSecret(ctx, cluster.ObjectMeta)
		if IsCredentialsNotFoundError(err) {
			credentialSecret, err = u.getLegacyCredentialSecret(ctx, cluster.ObjectMeta)
			if IsCredentialsNotFoundError(err) {
				credentialSecret = &v1.Secret{}
				err := u.ctrlClient.Get(ctx, ctrlclient.ObjectKey{Namespace: credentialDefaultNamespace, Name: credentialDefaultName}, credentialSecret)
				if err != nil {
					u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Skipping Cluster %#q", cluster.Name), "stack", microerror.JSON(err))
					continue
				}
			} else if err != nil {
				u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Skipping Cluster %#q", cluster.Name), "stack", microerror.JSON(err))
			}
		} else if err != nil {
			u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Skipping Cluster %#q", cluster.Name), "stack", microerror.JSON(err))
		}

		clustersSecret[cluster.Name] = credentialSecret
	}

	return clustersSecret, nil
}

// getOrganizationCredentialSecret tries to find a Secret in the organization namespace.
func (u *VMSSRateLimit) getOrganizationCredentialSecret(ctx context.Context, objectMeta metav1.ObjectMeta) (*v1.Secret, error) {
	organization, ok := objectMeta.Labels[label.Organization]
	if !ok {
		return nil, missingOrganizationLabel
	}

	secretList := &v1.SecretList{}
	{
		err := u.ctrlClient.List(
			ctx,
			secretList,
			ctrlclient.InNamespace(objectMeta.Namespace),
			ctrlclient.MatchingLabels{
				"app":              "credentiald",
				label.Organization: organization,
			},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We currently only support one credential secret per organization.
	// If there are more than one, return an error.
	if len(secretList.Items) > 1 {
		return nil, microerror.Mask(tooManyCredentialsError)
	}

	if len(secretList.Items) < 1 {
		return nil, microerror.Mask(credentialsNotFoundError)
	}

	// If one credential secret is found, we use that.
	return &secretList.Items[0], nil
}

// getLegacyCredentialSecret tries to find a Secret in the default credentials namespace but labeled with the organization name.
// This is needed while we migrate everything to the org namespace and org credentials are created in the org namespace instead of the default namespace.
func (u *VMSSRateLimit) getLegacyCredentialSecret(ctx context.Context, objectMeta metav1.ObjectMeta) (*v1.Secret, error) {
	organization, ok := objectMeta.Labels[label.Organization]
	if !ok {
		return nil, missingOrganizationLabel
	}

	secretList := &v1.SecretList{}
	{
		err := u.ctrlClient.List(
			ctx,
			secretList,
			ctrlclient.InNamespace(credentialDefaultNamespace),
			ctrlclient.MatchingLabels{
				"app":              "credentiald",
				label.Organization: organization,
			},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We currently only support one credential secret per organization.
	// If there are more than one, return an error.
	if len(secretList.Items) > 1 {
		return nil, microerror.Mask(tooManyCredentialsError)
	}

	if len(secretList.Items) < 1 {
		return nil, microerror.Mask(credentialsNotFoundError)
	}

	// If one credential secret is found, we use that.
	return &secretList.Items[0], nil
}

// collectMeasuredCallsFromResponse When being throttled, the response will contain information with the number of calls being made.
// https://docs.microsoft.com/en-us/azure/virtual-machines/troubleshooting/troubleshooting-throttling-errors#throttling-error-details
func (u *VMSSRateLimit) collectMeasuredCallsFromResponse(ch chan<- prometheus.Metric, result compute.VirtualMachineScaleSetVMListResultIterator, subscriptionID, clientID string) {
	data := tryParseRequestCountFromResponse(result.Response().Response)
	for k, v := range data {
		ch <- prometheus.MustNewConstMetric(
			vmssMeasuredCallsDesc,
			prometheus.GaugeValue,
			v,
			subscriptionID,
			clientID,
			k,
		)
	}
}

func (u *VMSSRateLimit) Describe(ch chan<- *prometheus.Desc) error {
	ch <- vmssVMListDesc
	ch <- vmssMeasuredCallsDesc
	return nil
}

func inArray(a []string, s string) bool {
	for _, x := range a {
		if x == s {
			return true
		}
	}

	return false
}

// This function is a best-effort attempt at reading the number of API calls we are making
// towards the Azure VMSS API during a 429.
// Useful metric to check if the situation is improving or not.
func tryParseRequestCountFromResponse(response autorest.Response) map[string]float64 {
	ret := map[string]float64{}

	type detail struct {
		Message string `json:"message"`
	}

	type azureerr struct {
		Details []detail `json:"details"`
	}

	type errorbody struct {
		Error azureerr `json:"error"`
	}

	var azz errorbody
	d := json.NewDecoder(response.Body)

	err := d.Decode(&azz)
	if err != nil {
		return ret
	}

	// {"operationGroup":"HighCostGetVMScaleSet30Min","startTime":"2020-10-05T14:33:39.6092603+00:00","endTime":"2020-10-05T14:50:00+00:00","allowedRequestCount":937,"measuredRequestCount":3277}

	type msg struct {
		OperationGroup       string `json:"operationGroup"`
		MeasuredRequestCount int64  `json:"measuredRequestCount"`
	}

	for _, m := range azz.Error.Details {
		var k msg
		err = json.Unmarshal([]byte(m.Message), &k)
		if err != nil {
			return ret
		}

		ret[k.OperationGroup] = float64(k.MeasuredRequestCount)
	}

	return ret
}
