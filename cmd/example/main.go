package main

import (
	"context"
	"net/http"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/service"

	"golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/gen/schema"
	"golang.cisco.com/examples/example/pkg/handlers"
)

func GETOverride(ctx context.Context, event *examplev1.AgentDbReadEvent) (examplev1.Agent, int, error) {
	payloadObject := event.Resource().(examplev1.Agent)
	result := examplev1.AgentFactory()
	desc := payloadObject.Spec().Description()
	id := payloadObject.Spec().ID()
	agentPl := payloadObject.Spec().AgentPool()
	org := payloadObject.Spec().Organization()
	name := payloadObject.Spec().Name()
	token := "******"
	if err := core.NewError(result.SpecMutable().SetDescription(desc),
		result.SpecMutable().SetID(id), result.SpecMutable().SetAgentPool(agentPl),
		result.SpecMutable().SetOrganization(org), result.SpecMutable().SetName(name),
		result.SpecMutable().SetToken(token)); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func onStart(ctx context.Context, changer mo.Changer) error {
	log := core.LoggerFromContext(ctx)

	log.Info("agent service start")
	return nil
}

func main() {
	handlerReg := []interface{}{
		handlers.AgentHandler,
	}
	examplev1.AgentMeta().RegisterAPIMethodGET(GETOverride)
	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlerReg...); err != nil {
		panic(err)
	}
}
