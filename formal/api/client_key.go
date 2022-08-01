package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type KeyResponse struct {
	Key KeyStruct `json:"key"`
}

const keyApiPath = "/admin/integrations/encryption"

// Done 2
func (c *Client) CreateKey(payload KeyStruct) (*KeyStruct, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+ keyApiPath, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	createdKeyRes := KeyResponse{}
	err = json.Unmarshal(body, &createdKeyRes)

	if err != nil {
		return nil, err
	}

	return &createdKeyRes.Key, nil
}

// Done 2
// GetKey - Returns a specifc key
func (c *Client) GetKey(formalKeyId string) (*KeyStruct, error) {
	req, err := http.NewRequest("GET", c.HostURL+keyApiPath + "/" + formalKeyId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	getKeyRes := KeyResponse{}
	err = json.Unmarshal(body, &getKeyRes)
	if err != nil {
		return nil, err
	}


	return &getKeyRes.Key, nil
}


func (c *Client) DeleteKey(formalKeyId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+keyApiPath + "/" + formalKeyId, nil)

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
