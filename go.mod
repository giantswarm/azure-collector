module github.com/giantswarm/azure-collector/v3

go 1.18

require (
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.29
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.12
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/giantswarm/apiextensions/v6 v6.6.0
	github.com/giantswarm/exporterkit v1.1.0
	github.com/giantswarm/k8sclient/v7 v7.2.0
	github.com/giantswarm/microendpoint v1.1.0
	github.com/giantswarm/microerror v0.4.1
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v1.1.1
	github.com/giantswarm/operatorkit/v2 v2.0.2
	github.com/giantswarm/statusresource/v5 v5.0.0
	github.com/giantswarm/versionbundle v1.1.0
	github.com/google/go-cmp v0.6.0
	github.com/prometheus/client_golang v1.17.0
	github.com/spf13/viper v1.17.0
	golang.org/x/sync v0.5.0
	k8s.io/api v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v1.5.2
	sigs.k8s.io/cluster-api v1.3.5
	sigs.k8s.io/cluster-api-provider-azure v1.8.2
	sigs.k8s.io/controller-runtime v0.16.3
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.5.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.2.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.9.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/giantswarm/backoff v1.0.0 // indirect
	github.com/giantswarm/certs/v3 v3.1.1 // indirect
	github.com/giantswarm/errors v0.3.0 // indirect
	github.com/giantswarm/k8sclient/v5 v5.12.0 // indirect
	github.com/giantswarm/operatorkit/v7 v7.1.0 // indirect
	github.com/giantswarm/tenantcluster/v4 v4.1.0 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gobuffalo/flect v1.0.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jongio/azidext/go/azidext v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/gomega v1.27.6 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/sagikazarmark/locafero v0.3.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.15.0 // indirect
	golang.org/x/exp v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/term v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.28.3 // indirect
	k8s.io/component-base v0.26.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/gin-gonic/gin v1.4.0 => github.com/gin-gonic/gin v1.9.0
	github.com/labstack/echo/v4 v4.1.11 => github.com/labstack/echo/v4 v4.10.2
	github.com/microcosm-cc/bluemonday v1.0.2 => github.com/microcosm-cc/bluemonday v1.0.23
	github.com/nats-io/nats-server/v2 v2.5.0 => github.com/nats-io/nats-server/v2 v2.9.15
	github.com/valyala/fasthttp v1.6.0 => github.com/valyala/fasthttp v1.45.0
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.12.0
	k8s.io/api => k8s.io/api v0.25.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.25.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.25.0
	k8s.io/client-go => k8s.io/client-go v0.25.0
	k8s.io/component-base => k8s.io/component-base v0.25.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.3.5
	sigs.k8s.io/cluster-api-provider-azure => sigs.k8s.io/cluster-api-provider-azure v1.8.2
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.13.1
)
