package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Done 2
// CreatePolicyLink - Create new policyLink
type CreatePolicyLinkResponseForGroup struct {
	Message     string             `json:"message"`
	PolicyLinks []PolicyLinkStruct `json:"policy_links"`
}
type CreatePolicyLinkResponseForRoleAndDatastore struct {
	Message    string           `json:"message"`
	PolicyLink PolicyLinkStruct `json:"policy_link"`
}

func (c *Client) CreatePolicyLink(policyPayload PolicyLinkStruct) (*PolicyLinkStruct, error) {
	itemId := policyPayload.ItemID

	// Compile
	postUrl := "/admin"
	var routePayload interface{}
	if policyPayload.Type == "group" {
		postUrl += "/identities/groups/" + itemId + "/link/policies"
		routePayload = struct {
			Policies []string `json:"policies"`
		}{
			Policies: []string{policyPayload.PolicyID},
		}
	}
	if policyPayload.Type == "role" {
		postUrl += "/identities/roles/link/" + itemId + "?policy_id=" + policyPayload.PolicyID
	}
	if policyPayload.Type == "datastore" {
		postUrl += "/stores/" + itemId + "/link" + "?policy_id=" + policyPayload.PolicyID
	}

	// Send
	rb, err := json.Marshal(routePayload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.HostURL+postUrl, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var createdPolicyLink PolicyLinkStruct
	if policyPayload.Type == "group" {
		var ret CreatePolicyLinkResponseForGroup
		err = json.Unmarshal(body, &ret)
		if err != nil {
			return nil, err
		}
		if len(ret.PolicyLinks) != 1 {
			return nil, errors.New("the newly created policy was not received. An Internal Server Error occurred")
		}
		createdPolicyLink = ret.PolicyLinks[0]
	}
	if policyPayload.Type == "role" {
		var ret CreatePolicyLinkResponseForRoleAndDatastore
		err = json.Unmarshal(body, &ret)
		if err != nil {
			return nil, err
		}
		createdPolicyLink = ret.PolicyLink
	}
	if policyPayload.Type == "datastore" {
		var ret CreatePolicyLinkResponseForRoleAndDatastore
		err = json.Unmarshal(body, &ret)
		if err != nil {
			return nil, err
		}
		createdPolicyLink = ret.PolicyLink
	}

	return &createdPolicyLink, nil
}

// Done 2
// GetPolicyLink - Returns a specifc policyLink
type GetPolicyLinkResponse struct {
	Message    string           `json:"message"`
	PolicyLink PolicyLinkStruct `json:"policy_link"`
}

func (c *Client) GetPolicyLink(policyLinkId string) (*PolicyLinkStruct, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/policies/links/"+policyLinkId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	policyLink := GetPolicyLinkResponse{}
	err = json.Unmarshal(body, &policyLink)
	if err != nil {
		return nil, err
	}

	return &policyLink.PolicyLink, nil
}

// UpdatePolicyLink - Updates an policyLink
// func (c *Client) UpdatePolicyLink(policyLinkId string, policyLinkUpdate PolicyLinkStruct) error {
// rb, err := json.Marshal(policyLinkUpdate)
// if err != nil {
// 	return err
// }

// req, err := http.NewRequest("PUT", c.HostURL+"/admin/policies/"+policyLinkId, strings.NewReader(string(rb)))
// if err != nil {
// 	return err
// }

// // TODO: Though the api restricts fields, best to restrict here as well
// body, err := c.doRequest(req)
// if err != nil {
// 	return err
// }

// var res Message
// err = json.Unmarshal(body, &res)
// if err != nil {
// 	return err
// }

// return nil
// }

// DeletePolicyLink - Deletes a policyLink
type DeleteRes struct {
	Message string `json:"message"`
}
func (c *Client) DeletePolicyLink(ctx context.Context, policyLinkId string) error {
	existingPolicyLink, err := c.GetPolicyLink(policyLinkId)
	if err != nil {
		return err
	}
	deleteUrl := "/admin"
	if existingPolicyLink.Type == "group" {
		deleteUrl += "/identities/groups/" + existingPolicyLink.ItemID + "/link/policies/" + existingPolicyLink.PolicyID
	}
	if existingPolicyLink.Type == "role" {
		deleteUrl += "/identities/roles/link/" + existingPolicyLink.ItemID + "/link/" + existingPolicyLink.PolicyID
	}
	if existingPolicyLink.Type == "datastore" {
		// Done
		deleteUrl += "/stores/" + existingPolicyLink.ItemID + "/link/" + existingPolicyLink.PolicyID
	}

	req, err := http.NewRequest("DELETE", c.HostURL+deleteUrl, nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	var res DeleteRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	return nil
}
