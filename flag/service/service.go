package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/azure-collector/flag/service/azure"
)

type Service struct {
	Azure      azure.Azure
	Kubernetes kubernetes.Kubernetes
}
