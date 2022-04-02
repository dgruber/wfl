module github.com/dgruber/wfl

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.11
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
)

require (
	github.com/deepmap/oapi-codegen v1.9.1
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/drmaa2os v0.3.20
	github.com/go-chi/chi/v5 v5.0.7
	github.com/mitchellh/copystructure v1.2.0
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.19.0
	github.com/rs/zerolog v1.26.1
	github.com/sirupsen/logrus v1.8.1
	k8s.io/klog/v2 v2.60.1
)

require (
	cloud.google.com/go v0.81.0 // indirect
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f // indirect
	code.cloudfoundry.org/lager v2.0.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/cloudfoundry-community/go-cfclient v0.0.0-20220207220839-752842e14060 // indirect
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgruber/drmaa v1.0.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.14+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/getkin/kin-openapi v0.93.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.22.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.0.0-20211215165025-cf75a172585e // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211005153810-c76a74d43a8e // indirect
	google.golang.org/grpc v1.41.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.22.6 // indirect
	k8s.io/apimachinery v0.22.6 // indirect
	k8s.io/client-go v0.22.6 // indirect
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

go 1.17
