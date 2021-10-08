module golang.cisco.com/examples/example

go 1.15

replace golang.cisco.com/argo => ../argo

require (
	github.com/go-logr/zapr v0.4.0
	github.com/golangci/golangci-lint v1.42.1
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/prometheus/common v0.14.0 // indirect
	go.uber.org/zap v1.17.0
	golang.cisco.com/argo v0.0.0-00010101000000-000000000000
)
