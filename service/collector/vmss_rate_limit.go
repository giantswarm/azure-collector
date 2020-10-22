package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/azure-collector/v2/service/credential"
)

const (
	vmssVMListHeaderName = "X-Ms-Ratelimit-Remaining-Resource"
	vmssMetricsSubsystem = "rate_limit"
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
	G8sClient  versioned.Interface
	K8sClient  kubernetes.Interface
	Location   string
	Logger     micrologger.Logger
	GSTenantID string
}

type VMSSRateLimit struct {
	g8sClient  versioned.Interface
	k8sClient  kubernetes.Interface
	location   string
	logger     micrologger.Logger
	gsTenantID string
}

func init() {
	prometheus.MustRegister(vmssVMListErrorCounter)
}

func NewVMSSRateLimit(config VMSSRateLimitConfig) (*VMSSRateLimit, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.GSTenantID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GSTenantID must not be empty", config)
	}

	u := &VMSSRateLimit{
		g8sClient:  config.G8sClient,
		k8sClient:  config.K8sClient,
		location:   config.Location,
		logger:     config.Logger,
		gsTenantID: config.GSTenantID,
	}

	return u, nil
}

func (u *VMSSRateLimit) getResourceGroupName() string {
	return fmt.Sprintf("%s-%s", resourceGroupNamePrefix, u.location)
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

	azureClients, err := credential.GetAzureClientSetsFromCredentialSecrets(ctx, u.k8sClient, u.gsTenantID)
	if err != nil {
		return microerror.Mask(err)
	}

	for azureClientSetConfig, azureClientSet := range azureClients {
		var headers []string

		result, err := azureClientSet.VirtualMachineScaleSetsClient.ListComplete(ctx, u.getResourceGroupName())
		if err != nil {
			u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Error calling azure API"), "stack", microerror.JSON(err))
			detailed, ok := err.(autorest.DetailedError)
			if !ok {
				u.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Skipping clientid %#q / subscriptiondid %#q due to error calling Azure API", azureClientSetConfig.ClientID, azureClientSetConfig.SubscriptionID))
				continue
			}
			headers = detailed.Response.Header[vmssVMListHeaderName]

			data := tryParseRequestCountFromResponse(detailed)
			for k, v := range data {
				ch <- prometheus.MustNewConstMetric(
					vmssMeasuredCallsDesc,
					prometheus.GaugeValue,
					v,
					azureClientSetConfig.SubscriptionID,
					azureClientSetConfig.ClientID,
					k,
				)
			}
		}

		if len(headers) == 0 {
			headers = result.Response().Response.Header[vmssVMListHeaderName]
		}

		// Header not found, we consider this an error.
		if len(headers) == 0 {
			vmssVMListErrorCounter.Inc()
			continue
		}

		for _, l := range headers {
			// Limits are a single comma separated string.
			tokens := strings.SplitN(l, ",", -1)
			for _, t := range tokens {
				// Each limit's name and value are separated by a semicolon.
				kv := strings.SplitN(t, ";", 2)
				if len(kv) != 2 {
					// We expect exactly two tokens, otherwise we consider this a parsing error.
					vmssVMListErrorCounter.Inc()
					continue
				}

				// The second token must be a number or we don't know what we got from MS.
				val, err := strconv.ParseFloat(kv[1], 64)
				if err != nil {
					vmssVMListErrorCounter.Inc()
					continue
				}

				ch <- prometheus.MustNewConstMetric(
					vmssVMListDesc,
					prometheus.GaugeValue,
					val,
					azureClientSetConfig.SubscriptionID,
					azureClientSetConfig.ClientID,
					kv[0],
				)
			}
		}
	}

	return nil
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
func tryParseRequestCountFromResponse(detailed autorest.DetailedError) map[string]float64 {
	ret := map[string]float64{}

	body := detailed.Response.Body

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
	d := json.NewDecoder(body)

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
