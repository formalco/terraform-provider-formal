package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// CreatePolicy - Create new policy
func (c *Client) CreatePolicy(ctx context.Context, payload CreatePolicyPayload) (*PolicyOrgItem, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/policies", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	createdPolicy := PolicyOrgItem{}
	err = json.Unmarshal(body, &createdPolicy)
	if err != nil {
		return nil, err
	}

	return &createdPolicy, nil

	// createdPolicyRes := GetAndCreatePolicyEndpointRes{}
	// err = json.Unmarshal(body, &createdPolicyRes)
	// if err != nil {
	// 	return nil, err
	// }
	// return &createdPolicyRes.Policy, nil
}

// At the moment, only GET is shaped such. Create needs to be updated (is just struct atm)
type GetAndCreatePolicyEndpointRes struct {
	Policy PolicyOrgItem `json:"policy"`
}

// GetPolicy - Returns a specifc policy
func (c *Client) GetPolicy(policyId string) (*PolicyOrgItem, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/policies/"+policyId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	resJson := GetAndCreatePolicyEndpointRes{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		return nil, err
	}
	policy := resJson.Policy

	return &policy, nil
}

// UpdatePolicy - Updates an policy
// func (c *Client) UpdatePolicy(policyId string, policyUpdate PolicyOrgItem) error {
// 	rb, err := json.Marshal(policyUpdate)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequest("PUT", c.HostURL+"/admin/policies/"+policyId, strings.NewReader(string(rb)))
// 	if err != nil {
// 		return err
// 	}

// 	// TODO: Though the api restricts fields, best to restrict here as well
// 	body, err := c.doRequest(req)
// 	if err != nil {
// 		return err
// 	}

// 	var res Message
// 	err = json.Unmarshal(body, &res)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// DeletePolicy - Deletes a policy
func (c *Client) DeletePolicy(policyId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/policies/"+policyId, nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	var res Message
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	return nil
}
