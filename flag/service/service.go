package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/azure-collector/flag/service/azure"
	"github.com/giantswarm/azure-collector/flag/service/installation"
	"github.com/giantswarm/azure-collector/flag/service/tenant"
)

type Service struct {
	Azure          azure.Azure
	Installation   installation.Installation
	Kubernetes     kubernetes.Kubernetes
	RegistryDomain string
	Tenant         tenant.Tenant
}
