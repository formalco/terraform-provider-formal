package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type CreateGroupEndpointRes struct {
	Group   Group  `json:"group"`
	Message string `json:"message"`
}

// Done 2
// Create new group
func (c *Client) CreateGroup(payload Group) (*Group, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/identities/groups", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	createdGroupRes := CreateGroupEndpointRes{}
	err = json.Unmarshal(body, &createdGroupRes)

	if err != nil {
		return nil, err
	}

	return &createdGroupRes.Group, nil
}

type GetGroupEndpointRes struct {
	Group Group `json:"group"`
}

// Done 2
// GetGroup - Returns a specifc group
func (c *Client) GetGroup(groupId string) (*Group, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/identities/groups/"+groupId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	resJson := GetGroupEndpointRes{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		return nil, err
	}
	group := resJson.Group

	return &group, nil
}

// UpdateGroup - Updates an group
func (c *Client) UpdateGroup(groupId string, groupUpdate Group) error {
	rb, err := json.Marshal(groupUpdate)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/identities/groups/"+groupId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	// TODO: Though the api restricts fields, best to restrict here as well
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

// DeleteGroup - Deletes a group
func (c *Client) DeleteGroup(groupId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/identities/groups/"+groupId, nil)
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
