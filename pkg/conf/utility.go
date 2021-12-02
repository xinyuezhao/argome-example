package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/hashicorp/go-tfe"
	"golang.cisco.com/argo/pkg/core"
)

func QueryAgents(ctx context.Context, client *http.Client, tfeClient *tfe.Client, agentplId string) ([]Agent, error) {
	log := core.LoggerFromContext(ctx)
	// query agents in given agentpool
	url := fmt.Sprintf("https://app.terraform.io/api/v2/agent-pools/%s/agents", agentplId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf("Bearer %s", Usertoken)
	req.Header.Set("Authorization", auth)
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	b, err := httputil.DumpResponse(resp, true)
	log.Info("parsing response data")
	if err != nil {
		return nil, err
	}
	log.Info("response " + string(b))
	log.Info("respose status " + resp.Status)
	if resp.StatusCode != 200 {
		err := core.NewError(fmt.Errorf("error! Response content: %s", string(b)))
		return nil, err
	}
	result := Agents{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func QueryAgentId(ctx context.Context, agents []Agent, name string) string {
	log := core.LoggerFromContext(ctx)
	for _, agent := range agents {
		log.Info(fmt.Sprintf("agent queried %s, status: %s", agent.Attributes.Name, agent.Attributes.Status))
		if agent.Attributes.Name == name && agent.Attributes.Status == "idle" {
			return agent.Id
		}
	}
	return ""
}

func QueryFeatures(ctx context.Context, client *http.Client) (Feature, error) {
	log := core.LoggerFromContext(ctx)
	result := Feature{}
	// req, err := http.NewRequest(http.MethodGet, "http://localhost:9090/api/config/dn/appinstances/cisco-argome", nil)
	req, err := http.NewRequest(http.MethodGet, "https://10.23.248.65/api/config/dn/appinstances/cisco-argome", nil)
	req.Header.Set("Cookie", Cookie)
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	b, err := httputil.DumpResponse(resp, true)
	log.Info("parsing response data")
	if err != nil {
		return result, err
	}
	log.Info("response " + string(b))
	log.Info("respose status " + resp.Status)
	if resp.StatusCode != 200 {
		err := core.NewError(fmt.Errorf("error! Response content: %s", string(b)))
		return result, err
	}

	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, http.StatusInternalServerError, err
	// }
	// log.Info("before log response body")
	// // // log.Info("respose body " + j.(string))
	// log.Info(string(bodyBytes))
	// log.Info("after log instance")

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	for _, feature := range result.Instances[0].Features {
		log.Info("feature instance " + feature.Instance)
		log.Info("feature status " + feature.OperState)
		log.Info("feature config name " + feature.ConfigParameters.Name)
		log.Info("feature config token " + feature.ConfigParameters.Token)
	}
	return result, nil
}
