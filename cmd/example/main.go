package main

import (
	"context"
	"fmt"
	"net/http"

	tfe "github.com/hashicorp/go-tfe"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/service"

	"golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/gen/schema"
	"golang.cisco.com/examples/example/pkg/handlers"
)

func configTFC() (context.Context, *tfe.Client, error) {
	config := &tfe.Config{
		Token: "ai1yMKOzv3Mptg.atlasv1.lOseEHJzlB49Vz0fXTlFUFRGGTuugiP3040sr1MGGOkHgRqzQ9FrpiUJzyTH1DzzFTM",
	}
	client, err := tfe.NewClient(config)
	if err != nil {
		return nil, nil, err
	}
	// Create a context
	ctxTfe := context.Background()
	return ctxTfe, client, nil
}

// Create a new agentPool for an organization
func createAgentPool(ctx context.Context, client *tfe.Client, orgName, agentPlName string) (*tfe.AgentPool, error) {
	createOptions := tfe.AgentPoolCreateOptions{Name: &agentPlName}
	agentPl, err := client.AgentPools.Create(ctx, orgName, createOptions)
	if err != nil {
		return nil, err
	}
	return agentPl, nil
}

func ListOverride(ctx context.Context, event *mo.TypeHandlerEvent) ([]examplev1.AgentPool, int, error) {
	log := core.LoggerFromContext(ctx)
	params := event.Params
	log.Info(fmt.Sprintf("Params len %v", len(params)))
	for key, value := range params {
		log.Info("show key " + key)
		log.Info("show value " + value.(string))
	}
	log.Info("register overriding LIST")

	return nil, http.StatusOK, nil
}

func GETOverride(ctx context.Context, event *examplev1.AgentPoolDbReadEvent) (examplev1.AgentPool, int, error) {
	log := core.LoggerFromContext(ctx)

	log.Info("register overriding GET")
	log.Info("show indentity " + event.ID())
	log.Info("show dn " + event.DN())
	log.Info("org name is " + event.Resource().(examplev1.AgentPool).Spec().Organization())
	log.Info("agentPl name is " + event.Resource().(examplev1.AgentPool).Spec().Name())
	return nil, http.StatusNotFound, nil
}

func POSTOverride(ctx context.Context, event *examplev1.AgentPoolDbCreateEvent) (examplev1.AgentPool, int, error) {
	log := core.LoggerFromContext(ctx)

	log.Info("register overriding POST")
	log.Info("show indentity " + event.ID())
	log.Info("show dn " + event.DN())
	payloadObject := event.Resource().(examplev1.AgentPool)
	orgName := payloadObject.Spec().Organization()
	agentName := payloadObject.Spec().Name()
	ctxTfe, client, err := configTFC()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	agentPl, err := createAgentPool(ctxTfe, client, orgName, agentName)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	result := examplev1.AgentPoolFactory()
	errs := make([]error, 0)
	errs = append(errs, result.SpecMutable().SetName(agentPl.Name),
		result.SpecMutable().SetOrganization(agentPl.Organization.Name),
		result.SpecMutable().SetID(agentPl.ID),
		result.SpecMutable().SetName(agentName))
	if err := core.NewError(errs...); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func onStart(ctx context.Context, changer mo.Changer) error {
	log := core.LoggerFromContext(ctx)

	log.Info("register overriding GET and List during app start")
	return nil
}

func main() {
	handlerReg := []interface{}{
		handlers.AgentPoolHandler,
	}
	examplev1.AgentPoolMeta().RegisterAPIMethodList(ListOverride)
	examplev1.AgentPoolMeta().RegisterAPIMethodGET(GETOverride)
	examplev1.AgentPoolMeta().RegisterAPIMethodPOST(POSTOverride)
	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlerReg...); err != nil {
		panic(err)
	}
}
