package auth

import (
	"github.com/giantswarm/azure-collector/flag/service/installation/tenant/kubernetes/api/auth/provider"
)

type Auth struct {
	Provider provider.Provider
}
