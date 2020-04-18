package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"
)

type Service struct {
	ControlPlaneResourceGroup string
	Kubernetes                kubernetes.Kubernetes
	Location                  string
}
