package api
import (
	"encoding/json"
	"net/http"
	"strings"
)

type CreateRoleEndpointRes struct {
	Role    Role   `json:"role"`
	Message string `json:"message"`
}

// Done 2
// Create new role
func (c *Client) CreateRole(payload Role) (*Role, error) {
	fullPayload := struct {
		Role Role `json:"role"`
	}{
		Role: payload,
	}

	rb, err := json.Marshal(fullPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/identities/roles", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	createdRoleRes := CreateRoleEndpointRes{}
	err = json.Unmarshal(body, &createdRoleRes)

	if err != nil {
		return nil, err
	}

	return &createdRoleRes.Role, nil
}

type GetRoleEndpointRes struct {
	Role Role `json:"role"`
}

// Done 2
// GetRole - Returns a specifc role
func (c *Client) GetRole(roleId string) (*Role, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/identities/roles/"+roleId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	resJson := GetRoleEndpointRes{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		return nil, err
	}
	role := resJson.Role

	return &role, nil
}

// UpdateGroup - Updates an group
func (c *Client) UpdateRole(roleId string, roleData Role) error {
	rb, err := json.Marshal(roleData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT",  c.HostURL+"/admin/identities/roles/"+roleId, strings.NewReader(string(rb)))
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


// DeleteRole - Deletes a role
func (c *Client) DeleteRole(roleId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/identities/roles/"+roleId, nil)
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
