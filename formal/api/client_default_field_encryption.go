package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type DefaultFieldEncryptionRes struct {
	DefaultFieldEncryption DefaultFieldEncryptionStruct `json:"default_field_encryption_policy"`
}

func (c *Client) CreateOrUpdateDefaultFieldEncryption(payload DefaultFieldEncryptionStruct) (*DefaultFieldEncryptionStruct, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/encryption/default-policy", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var createdDefaultFieldEncryption DefaultFieldEncryptionRes
	err = json.Unmarshal(body, &createdDefaultFieldEncryption)

	if err != nil {
		return nil, err
	}

	return &createdDefaultFieldEncryption.DefaultFieldEncryption, nil
}


// GetDefaultFieldEncryption - Returns a specifc defaultFieldEncryption
// Done 2
func (c *Client) GetDefaultFieldEncryption() (*DefaultFieldEncryptionStruct, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/encryption/default-policy", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	resJson := DefaultFieldEncryptionRes{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		return nil, err
	}
	

	return &resJson.DefaultFieldEncryption, nil
}

// DeleteDefaultFieldEncryption - Deletes a defaultFieldEncryption
// DONE
func (c *Client) DeleteDefaultFieldEncryption() error {
	req, err := http.NewRequest("DELETE",c.HostURL+"/admin/encryption/default-policy", nil)
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
