module github.com/dgruber/wfl/examples/kubernetes

go 1.15

replace (
	github.com/dgruber/wfl => ../../../wfl
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
)

require github.com/dgruber/wfl v0.3.7

