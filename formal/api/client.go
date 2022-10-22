package api

import (
	"errors"
	"fmt"
	"io"
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

const DEV_URL string = "http://localhost:4000"

// NewClient -
func NewClient(client_id, secret_key string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 100 * time.Second},
		ClientId:   client_id,
		SecretKey:  secret_key,
		HostURL:    FORMAL_HOST_URL,
	}

	if DEV_URL != "" {
		c.HostURL = DEV_URL
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if c.ClientId != "" && c.SecretKey != "" {
		req.Header.Add("client_id", c.ClientId)
		req.Header.Add("api_key", c.SecretKey)
	} else {
		return nil, errors.New("no client_id and api_key detected")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
