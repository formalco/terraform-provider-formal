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
	if dataplaneId == "" {
		return errors.New("dataplaneId is empty")
	}
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/integrations/cloud/stacks/"+dataplaneId, nil)

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

func (c *Client) CreateDataplaneRoutes(payload DataplaneRoutes) (*DataplaneRoutes, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/integrations/cloud/stacks/"+payload.DataplaneId+"/routes",
		strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dataplaneRoutes := DataplaneRoutes{}
	err = json.Unmarshal(body, &dataplaneRoutes)
	if err != nil {
		return nil, err
	}

	return &dataplaneRoutes, nil
}

func (c *Client) GetDataplaneRoutes(id string) (*DataplaneRoutes, error) {
	// Send GET request
	req, err := http.NewRequest("GET", c.HostURL+"/admin/integrations/cloud/stacks/routes/"+id, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Parse response
	routes := DataplaneRoutes{}
	err = json.Unmarshal(body, &routes)
	if err != nil {
		return nil, err
	}

	return &routes, nil
}

func (c *Client) DeleteDataplaneRoutes(routeId string) error {
	if routeId == "" {
		return errors.New("routeId is empty")
	}

	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/integrations/cloud/stacks/routes/"+routeId, nil)
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
