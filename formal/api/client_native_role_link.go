package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// CreateNativeRoleLink - Create new link from user to group
type NativeRoleLinkResponse struct {
	Link NativeRoleLink `json:"link"`
}

func (c *Client) CreateNativeRoleLink(datastoreId, nativeRoleId, formalIdentityId, formalIdentityType string) error {
	payload := NativeRoleLink{
		FormalIdentityType: formalIdentityType,
	}

	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/"+nativeRoleId+"/link/"+formalIdentityId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return err
	}
	var ret NativeRoleLinkResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetNativeRoleLink(datastoreId, identityId string) (*NativeRoleLink, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/identity-links/"+identityId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var ret NativeRoleLinkResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret.Link, nil
}

func (c *Client) DeleteNativeRoleLink(datastoreId, formalIdentityId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/stores/"+datastoreId+"/native-roles/ignored/link/"+formalIdentityId, nil)

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
