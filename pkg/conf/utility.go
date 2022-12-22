package conf

import (
	"bytes"
	"context"
	"crypto/tls"
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
	req, err := http.NewRequest(http.MethodGet, "https://10.23.248.67/api/config/dn/appinstances/cisco-argome", nil)
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

func ConfigTFC() (context.Context, *tfe.Client, error) {
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

func ConfigTLSClient(ctx context.Context) *http.Client {
	log := core.LoggerFromContext(ctx)
	log.Info("config TLS client")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return client
}

// Query agentPool by the name
func QueryAgentPlByName(agentPools []*tfe.AgentPool, name string) (*tfe.AgentPool, error) {
	for _, agentPl := range agentPools {
		if agentPl.Name == name {
			return agentPl, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("There is no agentPool named %v", name))
}

// Query all agentPools for an organization
func QueryAgentPools(ctx context.Context, client *tfe.Client, name string) ([]*tfe.AgentPool, error) {
	agentPools, err := client.AgentPools.List(ctx, name, tfe.AgentPoolListOptions{})
	if err != nil {
		return nil, err
	}
	res := agentPools.Items
	return res, nil
}

// Create a new agentToken
func CreateAgentToken(ctx context.Context, client *tfe.Client, agentPool, organization, desc string) (*tfe.AgentToken, string, error) {
	agentPools, _ := QueryAgentPools(ctx, client, organization)
	agentPl, queryErr := QueryAgentPlByName(agentPools, agentPool)
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
func RemoveAgentToken(ctx context.Context, client *tfe.Client, agentTokenID string) error {
	err := client.AgentTokens.Delete(ctx, agentTokenID)
	if err != nil {
		return err
	}
	return nil
}

// Delete an existing feature instance to stop agent
func DelFeatureInstance(ctx context.Context, client *http.Client, name string) error {
	log := core.LoggerFromContext(ctx)
	payload := map[string]string{
		"featureName": FeatureName,
		"app":         App,
		"instance":    name,
		"version":     Version,
		"vendor":      Vendor,
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(payload)
	req, err := http.NewRequest(http.MethodPost, "https://10.23.248.67/api/config/delfeatureinstance", payloadBuf)
	req.Header.Set("Cookie", Cookie)
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
