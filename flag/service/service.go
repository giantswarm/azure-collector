package service

import (
	"github.com/giantswarm/operatorkit/v2/pkg/flag/service/kubernetes"

	"github.com/giantswarm/azure-collector/v3/flag/service/azure"
)

type Service struct {
	Azure                     azure.Azure
	ControlPlaneResourceGroup string
	Kubernetes                kubernetes.Kubernetes
	Location                  string
}
