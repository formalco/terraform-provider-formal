package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	ClientId   string
	SecretKey  string
}

const FORMAL_HOST_URL string = "https://api.formalcloud.net"

// NewClient -
func NewClient(client_id, secret_key string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		ClientId:   client_id,
		SecretKey:  secret_key,
		HostURL:   FORMAL_HOST_URL,
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if c.ClientId != "" && c.SecretKey != "" {
		req.Header.Add("client_id", c.ClientId)
		req.Header.Add("api_key", c.SecretKey)
	}else{
		return nil, errors.New("no client_id and api_key detected")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

// -
