module golang.cisco.com/examples/example
go 1.15

replace golang.cisco.com/argo => ../argo

require (
  github.com/go-logr/zapr v0.4.0
  go.uber.org/zap v1.13.0
)
