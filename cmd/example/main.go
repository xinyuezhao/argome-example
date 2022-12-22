package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/hashicorp/go-tfe"
	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/service"

	"golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/gen/schema"
	"golang.cisco.com/examples/example/pkg/conf"
	"golang.cisco.com/examples/example/pkg/handlers"
)

func configTLSClient(ctx context.Context) *http.Client {
	log := core.LoggerFromContext(ctx)
	log.Info("config TLS client")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return client
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

// Query AgentTokens in an agentPool
func queryAgentTokens(ctx context.Context, client *tfe.Client, agentPlID string) ([]*tfe.AgentToken, error) {
	agentTokens, err := client.AgentTokens.List(ctx, agentPlID)
	if err != nil {
		return nil, err
	}
	res := agentTokens.Items
	return res, nil
}

func queryAgentStatus(ctx context.Context, agentId string) (string, error) {
	log := core.LoggerFromContext(ctx)
	client := configTLSClient(ctx)
	// query agents inside given agentpool
	log.Info("agent Id given " + agentId)
	url := fmt.Sprintf("https://app.terraform.io/api/v2/agents/%s", agentId)
	log.Info("query url " + url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	// use user token to access terraform cloud API
	req.Header.Set("Authorization", "Bearer ZCUWZISXNFtWIg.atlasv1.vty2xgI8e0zuvzwgM9INeLvus2WYZPz5uziE1YU0UB27RiIDNunkXjFYxjlm7fDZxMc")
	resp, e := client.Do(req)
	if e != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := httputil.DumpResponse(resp, true)
	log.Info("parsing response data")
	if err != nil {
		return "", err
	}
	log.Info("after parse response data")
	log.Info("response " + string(b))
	log.Info("respose status " + resp.Status)
	if resp.StatusCode != 200 {
		err := core.NewError(fmt.Errorf("there is an error. Response content: %s", string(b)))
		return "", err
	}
	agentObj := conf.AgentStatus{}
	log.Info("before parsing resp.Body")
	err = json.NewDecoder(resp.Body).Decode(&agentObj)
	if err != nil {
		return "", err
	}
	log.Info("after parsing resp.Body")
	return agentObj.Data.Attributes.Status, err
}

func GETOverride(ctx context.Context, event *examplev1.TokenListDbReadEvent) (examplev1.TokenList, int, error) {
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

func GETAgentOverride(ctx context.Context, event *examplev1.AgentDbReadEvent) (examplev1.Agent, int, error) {
	log := core.LoggerFromContext(ctx)
	name := event.Resource().(examplev1.Agent).Spec().Name()
	// id := event.Resource().(examplev1.Agent).Spec().AgentId()
	obj, err := event.Store().ResolveByName(ctx, examplev1.AgentDNForDefault(name))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	payloadObject := obj.(examplev1.Agent)
	if err := core.NewError(payloadObject.Spec().MutableAgentSpecV1Example().SetToken("********")); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// get agentPoolId
	agentPlId := payloadObject.Spec().AgentpoolId()
	TLSclient := configTLSClient(ctx)
	_, TFEclient, err := configTFC()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// get agentId by agentPoolId & agentName
	agents, err := conf.QueryAgents(ctx, TLSclient, TFEclient, agentPlId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	agentId := conf.QueryAgentId(ctx, agents, name)
	// call feature api to get status
	// localhost
	// https://10.23.248.67/api/config/dn/appinstances/cisco-argome
	// whether query feature instance operstate right after it was created?
	features, err := conf.QueryFeatures(ctx, TLSclient)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	for _, feature := range features.Instances[0].Features {
		if feature.Instance == name {
			payloadObject.SpecMutable().SetStatus(feature.OperState)
		}
	}
	// query status
	status := payloadObject.Spec().Status()
	if status == "Running" {
		if agentId != "" {
			log.Info("id used to query status " + agentId)
			status, err := queryAgentStatus(ctx, agentId)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			if err := core.NewError(payloadObject.Spec().MutableAgentSpecV1Example().SetStatus(status)); err != nil {
				return nil, http.StatusInternalServerError, err
			}
		}
	}
	return payloadObject, http.StatusOK, nil
}

func ListOverride(ctx context.Context, event *mo.TypeHandlerEvent) ([]examplev1.Agent, int, error) {
	log := core.LoggerFromContext(ctx)
	objs := event.Resolver.ResolveByKind(ctx, examplev1.AgentMeta().MetaKey())
	result := make([]examplev1.Agent, 0)
	TLSclient := conf.ConfigTLSClient(ctx)
	_, TFEclient, err := conf.ConfigTFC()
	if err != nil {
		log.Info("error during config TFC")
		return nil, http.StatusInternalServerError, err
	}
	features, err := conf.QueryFeatures(ctx, TLSclient)
	if err != nil {
		log.Info("error during querying features")
		return nil, http.StatusInternalServerError, err
	}
	for _, obj := range objs {
		payloadObject := obj.(examplev1.Agent)
		if err := core.NewError(payloadObject.SpecMutable().SetToken("********")); err != nil {
			log.Info("error during set token")
			return nil, http.StatusInternalServerError, err
		}
		// get agentPoolId
		agentPlId := payloadObject.Spec().AgentpoolId()
		// get agentId by agentPoolId & agentName
		agents, err := conf.QueryAgents(ctx, TLSclient, TFEclient, agentPlId)
		if err != nil {
			log.Info("error during querying agents")
			return nil, http.StatusInternalServerError, err
		}
		name := payloadObject.Spec().Name()
		agentId := conf.QueryAgentId(ctx, agents, name)
		for _, feature := range features.Instances[0].Features {
			if feature.Instance == name {
				payloadObject.SpecMutable().SetStatus(feature.OperState)
			}
		}
		status := payloadObject.Spec().Status()
		log.Info("status before query status by id " + status)
		if status == "Running" {
			if agentId != "" {
				log.Info("query agents' status " + agentId)
				status, err := conf.QueryAgentStatus(ctx, agentId)
				if err != nil {
					log.Info("error during querying agent status")
					return nil, http.StatusInternalServerError, err
				}
				if err := core.NewError(payloadObject.SpecMutable().SetStatus(status)); err != nil {
					log.Info("error during set status")
					return nil, http.StatusInternalServerError, err
				}
			}
		}
		result = append(result, payloadObject)
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
	examplev1.AgentMeta().RegisterAPIMethodGET(GETAgentOverride)
	examplev1.AgentMeta().RegisterAPIMethodList(ListOverride)

	if err := service.New("example", schema.Schema()).
		OnStart(onStart).
		Start(handlerReg...); err != nil {
		panic(err)
	}
}
