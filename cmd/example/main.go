package main

import (
	"context"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/service"

	"golang.cisco.com/examples/example/gen/schema"
	"golang.cisco.com/examples/example/pkg/handlers"
)

func onStart(ctx context.Context, changer mo.Changer) error {
	log := core.LoggerFromContext(ctx)

	log.Info("configuring some objects during app start")
	return nil
}

func main() {
	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlers.OrganizationHandler); err != nil {
		panic(err)
	}
}
