package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"golang.cisco.com/argo/pkg/core"
	"golang.cisco.com/argo/pkg/mo"
	"golang.cisco.com/argo/pkg/model"

	"golang.cisco.com/examples/example/gen/examplev1"
	"golang.cisco.com/examples/example/pkg/conf"
)

func AgentHandler(ctx context.Context, event mo.Event) error {
	log := core.LoggerFromContext(ctx)
	log.Info("handling Agent", "resource", event.Resource())
	agent := event.Resource().(examplev1.Agent)
	agentPl := agent.Spec().Agentpool()
	org := agent.Spec().Organization()
	name := agent.Spec().Name()
	ctxTfe, client, err := conf.ConfigTFC()
	if err != nil {
		return err
	}
	if event.Operation() == model.CREATE {
		// TODO: Add logic to set status. Currently set 'created' as default.
		if agent.Spec().Token() == "" {
			log.Info("create agent without token")
			agentToken, agentPlID, err := conf.CreateAgentToken(ctxTfe, client, agentPl, org, agent.Spec().Description())
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
			agentpools, err := conf.QueryAgentPools(ctxTfe, client, org)
			if err != nil {
				return err
			}
			agentpool, err := conf.QueryAgentPlByName(agentpools, agentPl)
			if err != nil {
				return err
			}
			if err := core.NewError(agent.SpecMutable().SetAgentpoolId(agentpool.ID)); err != nil {
				return err
			}
		}
		agent.SpecMutable().SetStatus("created")
		// api call creating feature instance to deploy agent
		TLSclient := conf.ConfigTLSClient(ctx)
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
		req, e := http.NewRequest(http.MethodPost, "https://10.23.248.67/api/config/createfeatureinstance", payloadBuf)
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
		TLSclient := conf.ConfigTLSClient(ctx)
		log.Info("before delete agent feature instance")
		err := conf.DelFeatureInstance(ctx, TLSclient, name)
		if err != nil {
			return err
		}
		log.Info("after deleting feature instance")
		log.Info("remove agentToken")
		time.Sleep(10 * time.Second)
		removeErr := conf.RemoveAgentToken(ctxTfe, client, tokenID)
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
