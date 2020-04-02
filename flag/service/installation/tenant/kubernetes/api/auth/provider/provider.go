package provider

import (
	"github.com/giantswarm/azure-collector/flag/service/installation/tenant/kubernetes/api/auth/provider/oidc"
)

type Provider struct {
	OIDC oidc.OIDC
}
