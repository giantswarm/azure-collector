package service

import (
	"github.com/giantswarm/operatorkit/v2/pkg/flag/service/kubernetes"
)

type Service struct {
	ControlPlaneResourceGroup string
	Kubernetes                kubernetes.Kubernetes
	Location                  string
}
