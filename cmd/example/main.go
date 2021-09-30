package main

import (
	"context"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/service"

	v1 "golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/gen/schema"
	"golang.cisco.com/examples/example/pkg/handlers"
)

func onStart(ctx context.Context, changer mo.Changer) error {
	log := core.LoggerFromContext(ctx)

	helloWorld := v1.WorldFactory()
	if err := helloWorld.SpecMutable().SetName("hello"); err != nil {
		return err
	}

	log.Info("configuring some objects during app start",
		"metaNames", helloWorld.MetaNames(),
		"object", helloWorld)

	return changer.Apply(ctx, helloWorld)
}

func main() {
	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlers.WorldHandler); err != nil {
		panic(err)
	}
}
