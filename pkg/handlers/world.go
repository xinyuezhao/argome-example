package handlers

import (
	"context"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"

	"golang.cisco.com/examples/example/gen/examplev1"
)

func WorldHandler(ctx context.Context, event mo.Event) error {
	log := core.LoggerFromContext(ctx)
	log.Info("handling world", "resource", event.Resource())
	world := event.Resource().(examplev1.World)
	if world.Spec().Description() == "" {
		newWorld := examplev1.WorldFactory()
		errs := make([]error, 0)
		errs = append(errs, world.SpecMutable().SetDescription("A fancy world"),
			newWorld.SpecMutable().SetName("parallel-"+world.Spec().Name()),
			newWorld.SpecMutable().SetDescription("A fancy parallel world"),
			world.Meta().MutableManagedObjectMetaV1Argo().SetStatus(mo.StatusModified))
		if err := core.NewError(errs...); err != nil {
			return err
		}
		if err := event.Store().Record(ctx, world); err != nil {
			return err
		}
		if err := event.Store().Record(ctx, newWorld); err != nil {
			return err
		}
		if err := event.Store().Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
