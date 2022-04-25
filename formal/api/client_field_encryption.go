package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type CreateFieldEncryptionRes struct {
	DatastoreID     string                `json:"datastore_id"`
	FieldEncryption FieldEncryptionStruct `json:"field"`
}

func (c *Client) CreateFieldEncryption(payload FieldEncryptionStruct) (*FieldEncryptionStruct, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/stores/"+payload.DsId+"/encryption/field", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var createdFieldEncryption CreateFieldEncryptionRes
	err = json.Unmarshal(body, &createdFieldEncryption)

	if err != nil {
		return nil, err
	}

	return &createdFieldEncryption.FieldEncryption, nil
}

type GetFieldEncryptionEndpointRes struct {
	FieldEncryptions []FieldEncryptionStruct `json:"fields"`
}

// GetFieldEncryption - Returns a specifc fieldEncryption
// Done 2
func (c *Client) GetFieldEncryption(dataStoreId, targetPath string) (*FieldEncryptionStruct, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/stores/"+dataStoreId+"/encryption/field", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Adapting to existing response format
	resJson := GetFieldEncryptionEndpointRes{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		return nil, err
	}
	for _, resFieldEncryption := range resJson.FieldEncryptions {
		if targetPath == resFieldEncryption.Path {
			return &resFieldEncryption, nil
		}
	}

	return nil, nil
}

// UpdateFieldEncryption - Updates an fieldEncryption
// func (c *Client) UpdateFieldEncryption(fieldEncryptionId string, fieldEncryptionUpdate FieldEncryptionOrgItem) error {
// 	rb, err := json.Marshal(fieldEncryptionUpdate)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequest("PUT", c.HostURL+"/admin/policies/"+fieldEncryptionId, strings.NewReader(string(rb)))
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

// DeleteFieldEncryption - Deletes a fieldEncryption
// DONE
func (c *Client) DeleteFieldEncryption(dataStoreId, formalKeyId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/stores/"+dataStoreId+"/encryption/field/"+formalKeyId, nil)
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
