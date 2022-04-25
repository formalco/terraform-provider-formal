package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// CreateGroupLinkRole - Create new link from user to group
type CreateGroupLinkRoleResponse struct {
	Message string `json:"message"`
}

type GroupLinkRolePayload struct {
	RoleIds []string `json:"roles"`
}

func (c *Client) CreateGroupLinkRole(roleId, groupId string) error {
	// Compile
	routePayload := GroupLinkRolePayload{
		RoleIds: []string{roleId},
	}

	// Send
	rb, err := json.Marshal(routePayload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.HostURL+"/admin/identities/groups/"+groupId+"/link/users", strings.NewReader(string(rb)))
	if err != nil {
		return err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return err
	}
	var ret CreateGroupLinkRoleResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetGroupLinkRole(roleId, groupId string) (string, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/identities/groups/"+groupId, nil)
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	group := GetGroupEndpointRes{}
	err = json.Unmarshal(body, &group)
	if err != nil {
		return "", err
	}
	for _, existingLinkRoleId := range group.Group.RolesIDs {
		if existingLinkRoleId == roleId {
			return roleId, nil
		}
	}

	return "", nil
}

// UpdateGroupLinkRole - Updates an roleLinkGroup
// func (c *Client) UpdateGroupLinkRole(roleLinkGroupId string, roleLinkGroupUpdate GroupLinkRoleStruct) error {
// rb, err := json.Marshal(roleLinkGroupUpdate)
// if err != nil {
// 	return err
// }

// req, err := http.NewRequest("PUT", c.HostURL+"/admin/policies/"+roleLinkGroupId, strings.NewReader(string(rb)))
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

// DeleteGroupLinkRole - Deletes a roleLinkGroup
func (c *Client) DeleteGroupLinkRole(roleId, groupId string) error {
	// Compile
	routePayload := GroupLinkRolePayload{
		RoleIds: []string{roleId},
	}

	// Send
	rb, err := json.Marshal(routePayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/identities/groups/"+groupId+"/link/users", strings.NewReader(string(rb)))

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
