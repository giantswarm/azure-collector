package azure

import (
	"github.com/giantswarm/azure-collector/flag/service/azure/hostcluster"
)

type Azure struct {
	ClientID        string
	ClientSecret    string
	EnvironmentName string
	HostCluster     hostcluster.HostCluster
	Location        string
	SubscriptionID  string
	TenantID        string
}
