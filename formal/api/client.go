package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HostURL    string
	HTTPClient *http.Client
	APIKey     string
}

const FORMAL_HOST_URL string = "https://api.formalcloud.net"

func NewClient(apiKey string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 100 * time.Second},
		APIKey:     apiKey,
		HostURL:    FORMAL_HOST_URL,
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if c.APIKey == "" {
		return nil, errors.New("client was not initialized with an api key")
	}
	req.Header.Add("X-Api-Key", c.APIKey)

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
