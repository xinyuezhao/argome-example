package handlers

import (
	"context"

	tfe "github.com/hashicorp/go-tfe"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"

	"golang.cisco.com/examples/example/gen/examplev1"
)

func AgentHandler(ctx context.Context, event mo.Event) error {
	log := core.LoggerFromContext(ctx)
	log.Info("handling agent", "resource", event.Resource())
	agent := event.Resource().(examplev1.Agent)

	if agent.Spec().Description() != "" {
		config := &tfe.Config{
			Token: "ai1yMKOzv3Mptg.atlasv1.lOseEHJzlB49Vz0fXTlFUFRGGTuugiP3040sr1MGGOkHgRqzQ9FrpiUJzyTH1DzzFTM",
		}

		client, err := tfe.NewClient(config)
		if err != nil {
			return err
		}

		// Create a context
		ctxTfe := context.Background()

		// Query all organizations
		orgs, err := queryAllOrgs(ctxTfe, client)
		if err != nil {
			return err
		}
		// Create a new agentPool
		orgName := orgs[0]
		agentPlName := "agentPl_" + agent.Spec().Description()
		agentPl, err := createAgentPool(ctxTfe, client, orgName, agentPlName)
		if err != nil {
			return err
		}
		agentToken, err := createAgentToken(ctxTfe, client, agentPl, agent.Spec().Description())
		if err != nil {
			return err
		}
		if err := core.NewError(agent.SpecMutable().SetToken(agentToken.Token)); err != nil {
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
	return nil
}

func queryAllOrgs(ctx context.Context, client *tfe.Client) ([]string, error) {
	var res []string
	orgs, err := client.Organizations.List(ctx, tfe.OrganizationListOptions{})
	if err != nil {
		return nil, err
	}
	// filter orgs by entitlement
	for _, element := range orgs.Items {
		entitlements, ers := client.Organizations.Entitlements(ctx, element.Name)
		if ers != nil {
			return nil, ers
		}
		if entitlements.Agents {
			res = append(res, element.Name)
		}
	}
	return res, nil
}

func createAgentPool(ctx context.Context, client *tfe.Client, orgName, agentPlName string) (*tfe.AgentPool, error) {
	createOptions := tfe.AgentPoolCreateOptions{Name: &agentPlName}
	agentPl, err := client.AgentPools.Create(ctx, orgName, createOptions)
	if err != nil {
		return nil, err
	}
	return agentPl, nil
}

func createAgentToken(ctx context.Context, client *tfe.Client, agentPl *tfe.AgentPool, desc string) (*tfe.AgentToken, error) {
	agentToken, err := client.AgentTokens.Generate(ctx, agentPl.ID, tfe.AgentTokenGenerateOptions{Description: &desc})
	if err != nil {
		return nil, err
	}
	return agentToken, nil
}
