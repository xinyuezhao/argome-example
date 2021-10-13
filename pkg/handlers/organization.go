package handlers

import (
	"context"

	tfe "github.com/hashicorp/go-tfe"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
)

func OrganizationHandler(ctx context.Context, event mo.Event) error {
	log := core.LoggerFromContext(ctx)
	log.Info("handling organization", "resource", event.Resource())

	// config := &tfe.Config{
	// 	Token: "ai1yMKOzv3Mptg.atlasv1.lOseEHJzlB49Vz0fXTlFUFRGGTuugiP3040sr1MGGOkHgRqzQ9FrpiUJzyTH1DzzFTM",
	// }

	// client, err := tfe.NewClient(config)
	// if err != nil {
	// 	return err
	// }

	// // Create a context
	// ctxTfe := context.Background()

	// // Query all organizations and filter orgs by entitlement
	// orgs, err := queryAllOrgs(ctxTfe, client)
	// if err != nil {
	// 	return err
	// }

	// errs := make([]error, 0)
	// for _, org := range orgs {
	// 	newOrg := examplev1.OrganizationFactory()
	// 	errs = append(errs, newOrg.SpecMutable().SetName(org.Name),
	// 		newOrg.SpecMutable().SetEmail(org.Email))
	// 	if err := event.Store().Record(ctx, newOrg); err != nil {
	// 		return err
	// 	}
	// }
	// if err := core.NewError(errs...); err != nil {
	// 	return err
	// }
	// if err := event.Store().Commit(ctx); err != nil {
	// 	return err
	// }

	return nil
}

func queryAllOrgs(ctx context.Context, client *tfe.Client) ([]*tfe.Organization, error) {
	var res []*tfe.Organization
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
			res = append(res, element)
		}
	}
	return res, nil
}

func OrganizationQuery() ([]*tfe.Organization, error) {
	config := &tfe.Config{
		Token: "ai1yMKOzv3Mptg.atlasv1.lOseEHJzlB49Vz0fXTlFUFRGGTuugiP3040sr1MGGOkHgRqzQ9FrpiUJzyTH1DzzFTM",
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Create a context
	ctxTfe := context.Background()
	orgs, err := queryAllOrgs(ctxTfe, client)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}
