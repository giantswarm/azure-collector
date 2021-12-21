module github.com/giantswarm/azure-collector/v2

go 1.15

require (
	github.com/Azure/azure-sdk-for-go v48.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.19
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/giantswarm/apiextensions/v2 v2.6.2
	github.com/giantswarm/apiextensions/v3 v3.32.0
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v4 v4.1.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v2 v2.0.2
	github.com/giantswarm/statusresource/v2 v2.0.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.5.6
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.8.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	k8s.io/api v0.18.19
	k8s.io/apimachinery v0.18.19
	k8s.io/client-go v0.18.19
	sigs.k8s.io/cluster-api v1.0.2
	sigs.k8s.io/cluster-api-provider-azure v1.1.0
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
)
