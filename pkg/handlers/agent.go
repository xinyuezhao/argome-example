package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/hashicorp/go-tfe"
	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/model"

	"golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/pkg/conf"
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

func configTLSClient(ctx context.Context) *http.Client {
	log := core.LoggerFromContext(ctx)
	log.Info("config TLS client")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return client
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

// Delete an existing feature instance to stop agent
func delFeatureInstance(ctx context.Context, client *http.Client, name string) error {
	log := core.LoggerFromContext(ctx)
	payload := map[string]string{
		"featureName": conf.FeatureName,
		"app":         conf.App,
		"instance":    name,
		"version":     conf.Version,
		"vendor":      conf.Vendor,
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(payload)
	req, err := http.NewRequest(http.MethodPost, "https://10.23.248.65/api/config/delfeatureinstance", payloadBuf)
	req.Header.Set("Cookie", conf.Cookie)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	b, err := httputil.DumpResponse(resp, true)
	log.Info(fmt.Sprintf("parsing response data %s", string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err := core.NewError(fmt.Errorf("error! Response content: %s", string(b)))
		return err
	}
	return nil
}

func AgentHandler(ctx context.Context, event mo.Event) error {
	log := core.LoggerFromContext(ctx)
	log.Info("handling Agent", "resource", event.Resource())
	agent := event.Resource().(examplev1.Agent)
	agentPl := agent.Spec().Agentpool()
	org := agent.Spec().Organization()
	name := agent.Spec().Name()
	ctxTfe, client, err := configTFC()
	if err != nil {
		return err
	}
	if event.Operation() == model.CREATE {
		// TODO: Add logic to set status. Currently set 'created' as default.
		if agent.Spec().Token() == "" {
			log.Info("create agent without token")
			agentToken, agentPlID, err := createAgentToken(ctxTfe, client, agentPl, org, agent.Spec().Description())
			if err != nil {
				return err
			}

			if err := core.NewError(agent.SpecMutable().SetToken(agentToken.Token),
				agent.SpecMutable().SetTokenId(agentToken.ID),
				agent.SpecMutable().SetAgentpoolId(agentPlID)); err != nil {
				return err
			}
		}
		token := agent.Spec().Token()
		if agent.Spec().AgentpoolId() == "" {
			agentpools, err := queryAgentPools(ctxTfe, client, org)
			if err != nil {
				return err
			}
			agentpool, err := queryAgentPlByName(agentpools, agentPl)
			if err != nil {
				return err
			}
			if err := core.NewError(agent.SpecMutable().SetAgentpoolId(agentpool.ID)); err != nil {
				return err
			}
		}
		agent.SpecMutable().SetStatus("created")
		// api call creating feature instance to deploy agent
		TLSclient := configTLSClient(ctx)
		param := map[string]string{"token": token, "name": name}
		body := map[string]interface{}{
			"vendor":           conf.Vendor,
			"version":          conf.Version,
			"app":              conf.App,
			"featureName":      conf.FeatureName,
			"instance":         name,
			"configParameters": param,
		}

		payloadBuf := new(bytes.Buffer)
		json.NewEncoder(payloadBuf).Encode(body)
		req, e := http.NewRequest(http.MethodPost, "https://10.23.248.65/api/config/createfeatureinstance", payloadBuf)
		// req, e := http.NewRequest(http.MethodPost, "http://localhost:9090/api/config/createfeatureinstance", payloadBuf)
		if e != nil {
			return e
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", conf.Cookie)
		log.Info("before request post")
		resp, e := TLSclient.Do(req)
		log.Info("after request post")
		if e != nil {
			return e
		}
		log.Info("err after request post")
		defer resp.Body.Close()
		// parse resp.Body
		b, err := httputil.DumpResponse(resp, true)
		log.Info("parsing response data")
		if err != nil {
			return err
		}
		log.Info("after parse response data")
		log.Info("response " + string(b))
		log.Info("respose status " + resp.Status)
		if resp.StatusCode != 200 {
			err := core.NewError(fmt.Errorf("there is an error. Response content: %s", string(b)))
			return err
		}

		if err := event.Store().Record(ctx, agent); err != nil {
			return err
		}
		if err := event.Store().Commit(ctx); err != nil {
			core.LoggerFromContext(ctx).Error(err, "failed to commit Agent")
			return err
		}
		// time.Sleep(8 * time.Second)
		// agents, err := queryAgents(ctx, TLSclient, client, agent.Spec().AgentpoolId())
		// if err != nil {
		// 	return err
		// }
		// agentId := queryAgentId(ctx, agents, name)
		// if err := core.NewError(agent.SpecMutable().SetAgentId(agentId)); err != nil {
		// 	return err
		// }

		// resp, err := http.Post("http://localhost:9090/api/config/createfeatureinstance", "application/json",
		// 	bytes.NewBuffer(json_data))
		// log.Info("after post")
		// var res map[string]interface{}
		// json.NewDecoder(resp.Body).Decode(&res)
		// log.Info("err decoding response")
		// restring := res["json"].(string)
		// log.Info("response decode " + restring)
	}

	if event.Operation() == model.DELETE {
		tokenID := agent.Spec().TokenId()
		// delete feature instance to stop the agent
		TLSclient := configTLSClient(ctx)
		log.Info("before delete agent feature instance")
		err := delFeatureInstance(ctx, TLSclient, name)
		if err != nil {
			return err
		}
		log.Info("after deleting feature instance")
		log.Info("remove agentToken")
		time.Sleep(10 * time.Second)
		removeErr := removeAgentToken(ctxTfe, client, tokenID)
		if removeErr != nil {
			return removeErr
		}
		log.Info("after removing agentToken")
	}
	return nil
}

func AgentValidator(ctx context.Context, event mo.Validation) error {
	log := core.LoggerFromContext(ctx)
	// event.Operation() description requied if agent without token
	log.Info("validate Agent", "resource", event.Resource())
	agent := event.Resource().(examplev1.Agent)
	desc := agent.Spec().Description()
	name := agent.Spec().Name()
	empty := ""
	if name == "" {
		empty = "name"
	}
	if desc == "" {
		empty = "description"
	}
	if empty != "" {
		err := core.NewError(fmt.Errorf("Agent %s can't be blank", empty))
		return err
	}
	return nil
}
