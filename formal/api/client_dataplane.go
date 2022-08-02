package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// CreateDataplane - Create new dataplane
func (c *Client) CreateDataplane(payload FlatDataplane) (*FlatDataplane, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/integrations/cloud/stacks/new-dataplane", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dataplane := FlatDataplane{}
	err = json.Unmarshal(body, &dataplane)
	if err != nil {
		return nil, err
	}

	return &dataplane, nil
}


func (c *Client) GetDataplane(dataplaneId string) (*FlatDataplane, error) {
	// Send GET request
	req, err := http.NewRequest("GET", c.HostURL+"/admin/integrations/cloud/stacks/"+dataplaneId, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Parse response
	plane := FlatDataplane{}
	err = json.Unmarshal(body, &plane)
	if err != nil {
		return nil, err
	}

	return &plane, nil
}

// DeleteDataplane - Deletes a dataplane
func (c *Client) DeleteDataplane(dataplaneId string) error {
	if dataplaneId == ""{
		return errors.New("dataplaneId is empty")
	}
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/integrations/cloud/stacks/" + dataplaneId, nil)

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

