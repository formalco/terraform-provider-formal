package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GetAndCreateSidecarResponseV2 struct {
	Id      string    `json:"id"`
	Sidecar SidecarV2 `json:"sidecar"`
}

// CreateSidecar - Create new sidecar
func (c *Client) CreateSidecar(payload SidecarV2) (string, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.HostURL+"/admin/sidecars", strings.NewReader(string(rb)))
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	sidecar := GetAndCreateSidecarResponseV2{}
	err = json.Unmarshal(body, &sidecar)
	if err != nil {
		return "", err
	}

	return sidecar.Id, nil
}

// GetSidecar - Returns a specifc sidecar
func (c *Client) GetSidecar(sidecarId string) (*SidecarV2, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/sidecars/"+sidecarId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dsInfra := GetAndCreateSidecarResponseV2{}
	err = json.Unmarshal(body, &dsInfra)
	if err != nil {
		return nil, err
	}

	return &dsInfra.Sidecar, nil
}

type GetSidecarTlsCertResponse struct {
	Secret string `json:"secret"`
}

func (c *Client) GetSidecarTlsCert(sidecarId string) (*string, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/sidecars/"+sidecarId+"/tlscert", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	tlsCertRes := GetSidecarTlsCertResponse{}
	err = json.Unmarshal(body, &tlsCertRes)
	if err != nil {
		return nil, err
	}

	return &tlsCertRes.Secret, nil
}

// UpdateSidecarName
func (c *Client) UpdateSidecarName(sidecarId string, sidecarUpdate SidecarV2) error {
	rb, err := json.Marshal(sidecarUpdate)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("PUT", c.HostURL+"/admin/sidecars/"+sidecarId+"/name", strings.NewReader(string(rb)))
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

func (c *Client) UpdateSidecarVersion(sidecarId, version string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/admin/sidecars/%s/version/%s", c.HostURL, sidecarId, version), nil)
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

func (c *Client) UpdateSidecarHostname(sidecarId, hostname string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/admin/sidecars/%s/sidecar-hostname?hostname=%s", c.HostURL, sidecarId, hostname), nil)
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

func (c *Client) UpdateSidecarGlobalKMSEncrypt(sidecarId string, sidecarUpdate SidecarV2) error {
	if sidecarUpdate.FullKMSDecryption {
		req, err := http.NewRequest("PUT", c.HostURL+"/admin/sidecars/"+sidecarId+"/kms-decrypt-policy?enable=true", nil)
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

// DeleteSidecar - Deletes a sidecar
func (c *Client) DeleteSidecar(sidecarId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+"/admin/sidecars/"+sidecarId, nil)
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
