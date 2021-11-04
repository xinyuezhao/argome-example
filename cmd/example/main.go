package main

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-tfe"
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

// Query AgentTokens in an agentPool
func queryAgentTokens(ctx context.Context, client *tfe.Client, agentPlID string) ([]*tfe.AgentToken, error) {
	agentTokens, err := client.AgentTokens.List(ctx, agentPlID)
	if err != nil {
		return nil, err
	}
	res := agentTokens.Items
	return res, nil
}

func GETOverride(ctx context.Context, event *examplev1.TokenListDbReadEvent) (examplev1.TokenList, int, error) {
	log := core.LoggerFromContext(ctx)
	log.Info("override GET for TokenList")
	payloadObject := event.Resource().(examplev1.TokenList)
	agentPlId := payloadObject.Spec().Agentpool()
	ctxTfe, client, err := configTFC()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	tokens, err := queryAgentTokens(ctxTfe, client, agentPlId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	result := examplev1.TokenListFactory()
	errs := make([]error, 0)
	for _, token := range tokens {
		errs = append(errs, result.SpecMutable().TokensAppendEl(token.ID))
	}
	errs = append(errs, result.SpecMutable().SetAgentpool(agentPlId))
	if err := core.NewError(errs...); err != nil {
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
		handlers.AgentValidator,
	}

	examplev1.TokenListMeta().RegisterAPIMethodGET(GETOverride)

	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlerReg...); err != nil {
		panic(err)
	}
}
