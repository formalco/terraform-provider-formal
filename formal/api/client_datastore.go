package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type GetAndCreateDataStoreResponseV2 struct {
	DataStoreId string      `json:"datastore_id"`
	DataStore   DatastoreV2 `json:"datastore"`
}

// CreateDatastore - Create new datastore
func (c *Client) CreateDatastore(payload DatastoreV2) (string, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/datastores", strings.NewReader(string(rb)))
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	datastore := GetAndCreateDataStoreResponseV2{}
	err = json.Unmarshal(body, &datastore)
	if err != nil {
		return "", err
	}

	return datastore.DataStoreId, nil
}

// GetDatastore - Returns a specifc datastore
func (c *Client) GetDatastore(datastoreId string) (*DatastoreV2, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/datastores/"+datastoreId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dsInfra := GetAndCreateDataStoreResponseV2{}
	err = json.Unmarshal(body, &dsInfra)
	if err != nil {
		return nil, err
	}

	return &dsInfra.DataStore, nil
}

// UpdateDatastoreName
func (c *Client) UpdateDatastoreName(datastoreId string, datastoreUpdate DatastoreV2) error {
	rb, err := json.Marshal(datastoreUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/datastores/"+datastoreId+"/name", strings.NewReader(string(rb)))
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
func (c *Client) UpdateDatastoreHealthCheckDbName(datastoreId string, datastoreUpdate DatastoreV2) error {
	rb, err := json.Marshal(datastoreUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/datastores/"+datastoreId+"/health-check-db-name", strings.NewReader(string(rb)))
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

// To be used in the future with other fields too
func (c *Client) UpdateDatastoreDefaultAcccessBehavior(datastoreId string, datastoreUpdate DatastoreV2) error {
	rb, err := json.Marshal(datastoreUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/datastores/"+datastoreId+"/default-access-behavior", strings.NewReader(string(rb)))
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
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/datastores/"+datastoreId, nil)
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
