package handlers

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"

	"golang.cisco.com/examples/example/gen/examplev1"
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

// Query agentPool by the name
func queryAgentPlByName(agentPools []*tfe.AgentPool, name string) (*tfe.AgentPool, error) {
	for _, agentPl := range agentPools {
		if agentPl.Name == name {
			return agentPl, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("There is no agentPool named %v", name))
}

// Query all agentPools for an organization
func queryAgentPools(ctx context.Context, client *tfe.Client, name string) ([]*tfe.AgentPool, error) {
	agentPools, err := client.AgentPools.List(ctx, name, tfe.AgentPoolListOptions{})
	if err != nil {
		return nil, err
	}
	res := agentPools.Items
	return res, nil
}

// Create a new agentToken
func createAgentToken(ctx context.Context, client *tfe.Client, agentPool, organization, desc string) (*tfe.AgentToken, string, error) {
	agentPools, _ := queryAgentPools(ctx, client, organization)
	agentPl, queryErr := queryAgentPlByName(agentPools, agentPool)
	if queryErr != nil {
		return nil, "", queryErr
	}
	agentToken, err := client.AgentTokens.Generate(ctx, agentPl.ID, tfe.AgentTokenGenerateOptions{Description: &desc})
	if err != nil {
		return nil, "", err
	}
	agentPlID := agentPl.ID
	return agentToken, agentPlID, nil
}

// Delete an existing agentToken
func removeAgentToken(ctx context.Context, client *tfe.Client, agentTokenID string) error {
	err := client.AgentTokens.Delete(ctx, agentTokenID)
	if err != nil {
		return err
	}
	return nil
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

// Delete an existing agentPool
func removeAgentPool(ctx context.Context, client *tfe.Client, agentPlID string) error {
	err := client.AgentPools.Delete(ctx, agentPlID)
	if err != nil {
		return err
	}
	return nil
}

func AgentHandler(ctx context.Context, event mo.Event) error {
	log := core.LoggerFromContext(ctx)
	log.Info("handling Agent", "resource", event.Resource())
	agent := event.Resource().(examplev1.Agent)
	log.Info("agent name given " + agent.Spec().Name())
	agentPl := agent.Spec().Agentpool()
	org := agent.Spec().Organization()
	operation := string(event.Operation())
	log.Info("operation is " + operation)
	ctxTfe, client, err := configTFC()
	if err != nil {
		return err
	}
	if operation == "CREATE" {
		log.Info("token is " + agent.Spec().Token())
		if agent.Spec().Token() == "" {
			log.Info("create agent without token")
			agentToken, agentPlID, err := createAgentToken(ctxTfe, client, agentPl, org, agent.Spec().Description())
			if err != nil {
				return err
			}

			if err := core.NewError(agent.SpecMutable().SetToken(agentToken.Token),
				agent.SpecMutable().SetId(agentToken.ID),
				agent.SpecMutable().SetAgentpoolId(agentPlID)); err != nil {
				return err
			}
			if err := event.Store().Record(ctx, agent); err != nil {
				return err
			}
			if err := event.Store().Commit(ctx); err != nil {
				core.LoggerFromContext(ctx).Error(err, "failed to commit Agent")
				return err
			}
		}
		log.Info("ID set " + agent.Spec().Id())
		log.Info("Token set " + agent.Spec().Token())
	}

	if operation == "DELETE" {
		tokenID := agent.Spec().Id()
		agentPools, _ := queryAgentPools(ctxTfe, client, org)
		agentPool, queryErr := queryAgentPlByName(agentPools, agentPl)
		if queryErr != nil {
			return queryErr
		}
		removeErr := removeAgentToken(ctxTfe, client, tokenID)
		if removeErr != nil {
			return removeErr
		}
		// check whether the AgentPool is empty, delete the AgentPool if yes
		tokensAfterDelete, err := queryAgentTokens(ctxTfe, client, agentPool.ID)
		if err != nil {
			return nil
		}
		if len(tokensAfterDelete) == 0 {
			removeEr := removeAgentPool(ctxTfe, client, agentPool.ID)
			if removeEr != nil {
				return removeEr
			}
		}
	}
	return nil
}
