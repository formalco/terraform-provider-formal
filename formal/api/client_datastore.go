package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type GetAndCreateDataStoreResponse struct {
	DataStoreId string         `json:"data_store_id"`
	DataStore   DataStoreInfra `json:"data_store"`
}

// CreateDatastore - Create new datastore
func (c *Client) CreateDatastore(payload DataStoreInfra) (*DataStoreInfra, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/stores", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	datastore := GetAndCreateDataStoreResponse{}
	err = json.Unmarshal(body, &datastore)
	if err != nil {
		return nil, err
	}

	return &datastore.DataStore, nil
}

// GetDatastore - Returns a specifc datastore
func (c *Client) GetDatastore(datastoreId string) (*DataStoreInfra, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/stores/"+datastoreId+"/infra", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dsInfra := GetAndCreateDataStoreResponse{}
	err = json.Unmarshal(body, &dsInfra)
	if err != nil {
		return nil, err
	}

	return &dsInfra.DataStore, nil
}


func (c *Client) UpdateDatastoreGlobalKMSEncrypt(datastoreId string, datastoreUpdate DataStoreInfra) error {
	if datastoreUpdate.FullKMSDecryption {
		req, err := http.NewRequest("PUT", c.HostURL+"/admin/stores/"+datastoreId+"/kms-decrypt-policy?enable=true", nil)
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

func (c *Client) UpdateDatastoreUsernamePassword(datastoreId string, datastoreUpdate DataStoreInfra) error {
	rb, err := json.Marshal(datastoreUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/stores/"+datastoreId+"/credentials", strings.NewReader(string(rb)))
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


// UpdateDatastoreName
func (c *Client) UpdateDatastoreName(datastoreId string, datastoreUpdate DataStoreInfra) error {
	rb, err := json.Marshal(datastoreUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/stores/"+datastoreId+"/name", strings.NewReader(string(rb)))
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

// UpdateDatastoreName
func (c *Client) UpdateDatastoreHealthCheckDbName(datastoreId string, datastoreUpdate DataStoreInfra) error {
	rb, err := json.Marshal(datastoreUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/stores/"+datastoreId+"/health-check-db-name", strings.NewReader(string(rb)))
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


// DeleteDatastore - Deletes a datastore
func (c *Client) DeleteDatastore(datastoreId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/stores/"+datastoreId, nil)
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

type GetDataStoreStatusResponse struct {
	DataStore DataStore `json:"data_store"`
}

// GetDatastore - Returns a specifc datastore
func (c *Client) GetDatastoreForStatus(datastoreId string) (*DataStore, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/stores/"+datastoreId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dsStatusRes := GetDataStoreStatusResponse{}
	err = json.Unmarshal(body, &dsStatusRes)
	if err != nil {
		return nil, err
	}

	return &dsStatusRes.DataStore, nil
}

type GetDataStoreTlsCertResponse struct {
	Secret string `json:"secret"`
}

// GetDatastoreTlsCert - Returns a specifc datastore
func (c *Client) GetDatastoreTlsCert(datastoreId string) (*string, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/stores/"+datastoreId+"/tlscert", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	tlsCertRes := GetDataStoreTlsCertResponse{}
	err = json.Unmarshal(body, &tlsCertRes)
	if err != nil {
		return nil, err
	}

	return &tlsCertRes.Secret, nil
}
