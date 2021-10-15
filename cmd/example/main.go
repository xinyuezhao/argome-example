package main

import (
	"context"
	"net/http"

	tfe "github.com/hashicorp/go-tfe"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/service"
	"golang.cisco.com/argo/pkg/utils"

	"golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/gen/schema"
	"golang.cisco.com/examples/example/pkg/handlers"
)

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

func newOrganization(org *tfe.Organization, newOrg examplev1.Organization) error {
	errs := make([]error, 0)
	errs = append(errs, newOrg.SpecMutable().SetName(org.Name),
		newOrg.SpecMutable().SetEmail(org.Email),
		newOrg.SpecMutable().SetCollaboratorAuthPolicy(string(org.CollaboratorAuthPolicy)),
		newOrg.SpecMutable().SetCostEstimationEnabled(org.CostEstimationEnabled),
		newOrg.SpecMutable().SetCreatedAt(org.CreatedAt.String()),
		newOrg.SpecMutable().SetExternalID(org.ExternalID),
		newOrg.SpecMutable().SetOwnersTeamSAMLRoleI(org.OwnersTeamSAMLRoleID),
		newOrg.SpecMutable().SetSAMLEnabled(org.SAMLEnabled),
		newOrg.SpecMutable().SetSessionRemember(org.SessionRemember),
		newOrg.SpecMutable().SetSessionTimeout(org.SessionTimeout),
		newOrg.SpecMutable().SetTrialExpiresAt(org.TrialExpiresAt.String()),
		newOrg.SpecMutable().SetTwoFactorConformant(org.TwoFactorConformant),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanCreateTeam(org.Permissions.CanCreateTeam),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanCreateWorkspace(org.Permissions.CanCreateWorkspace),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanCreateWorkspaceMigration(org.Permissions.CanCreateWorkspaceMigration),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanDestroy(org.Permissions.CanDestroy),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanTraverse(org.Permissions.CanTraverse),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanUpdate(org.Permissions.CanUpdate),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanUpdateAPIToken(org.Permissions.CanUpdateAPIToken),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanUpdateOAuth(org.Permissions.CanUpdateOAuth),
		newOrg.Spec().Permissions().MutableOrganizationPermissionsV1Example().
			SetCanUpdateSentinel(org.Permissions.CanUpdateSentinel))
	if err := core.NewError(errs...); err != nil {
		return err
	}
	return nil
}

func ListOverride(ctx context.Context, event *mo.TypeHandlerEvent) ([]examplev1.Organization, int, error) {
	ctxTfe, client, err := configTFC()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// Query all organizations and filter orgs by entitlement
	orgs, err := queryAllOrgs(ctxTfe, client)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	res := make([]examplev1.Organization, 0)
	for _, org := range orgs {
		newOrg := examplev1.OrganizationFactory()
		err := newOrganization(org, newOrg)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		res = append(res, newOrg)
	}
	return res, http.StatusOK, nil
}

func GETOverride(ctx context.Context, event *examplev1.OrganizationDbReadEvent) (examplev1.Organization, int, error) {
	payloadObject := event.Resource().(examplev1.Organization)
	name := payloadObject.Spec().Name()
	ctxTfe, client, err := configTFC()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	orgs, err := queryAllOrgs(ctxTfe, client)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for _, org := range orgs {
		if org.Name == name {
			newOrg := examplev1.OrganizationFactory()
			err := newOrganization(org, newOrg)
			if err == nil {
				return newOrg, http.StatusOK, nil
			}
		}
	}
	errMsg := utils.ErrorsNew(name + " not found")
	return nil, http.StatusNotFound, errMsg
}

func onStart(ctx context.Context, changer mo.Changer) error {
	log := core.LoggerFromContext(ctx)

	log.Info("register overriding GET and List during app start")
	return nil
}

func main() {
	handlerReg := []interface{}{
		handlers.OrganizationHandler,
	}
	examplev1.OrganizationMeta().RegisterAPIMethodList(ListOverride)
	examplev1.OrganizationMeta().RegisterAPIMethodGET(GETOverride)
	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlerReg...); err != nil {
		panic(err)
	}
}
