package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type NativeRoleRes struct {
	NativeRole NativeRole `json:"native_role"`
	Message    string     `json:"message"`
}

func (c *Client) CreateNativeRole(payload NativeRole) (*NativeRole, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/stores/"+payload.DatastoreId+"/native-roles", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	createdRoleRes := NativeRoleRes{}
	err = json.Unmarshal(body, &createdRoleRes)

	if err != nil {
		return nil, err
	}

	return &createdRoleRes.NativeRole, nil
}

func (c *Client) GetNativeRole(datastoreId, nativeRoleId string) (*NativeRole, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/"+nativeRoleId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	resJson := NativeRoleRes{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		return nil, err
	}
	role := resJson.NativeRole

	return &role, nil
}

type NewSecretPayload struct {
	Secret string `json:"secret"`
}

func (c *Client) UpdateNativeRole(datastoreId, roleId, newSecret string, useAsDefault bool) error {
	if newSecret != "" {
		payload := NewSecretPayload{Secret: newSecret}
		rb, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("PUT", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/"+roleId+"/secret", strings.NewReader(string(rb)))
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

	} else if useAsDefault {
		req, err := http.NewRequest("POST", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/"+roleId+"/default", nil)
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
	}

	return nil
}

// DeleteRole - Deletes a role
func (c *Client) DeleteNativeRole(datastoreId, roleId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/"+roleId, nil)
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
