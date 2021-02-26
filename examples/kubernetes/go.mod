module github.com/dgruber/wfl/examples/kubernetes

go 1.15

replace (
    github.com/dgruber/wfl => ../../../wfl
    github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.4
    k8s.io/api => k8s.io/api v0.20.2
    k8s.io/client-go => k8s.io/client-go v0.20.2
    k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
)

require (
	github.com/dgruber/drmaa2interface v1.0.2
	github.com/dgruber/wfl v0.3.8
)
