package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// APIEndpoints maps ralph-cli types to Ralph's API endpoints.
var APIEndpoints = map[string]string{
	"BaseObject":       "base-objects",
	"IPAddress":        "ipaddresses",
	"Ethernet":         "ethernets",
	"Memory":           "memory",
	"FibreChannelCard": "fibre-channel-cards",
	"Processor":        "processors",
	"Disk":             "disks",
	"DataCenterAsset":  "data-center-assets",
}

// Client provides an interface to interact with Ralph via its REST API.
type Client struct {
	scannedAddr Addr
	ralphURL    string
	apiKey      string
	apiVersion  string // Not used b/c Ralph doesn't have any API versioning (yet).
	client      *http.Client
}

// NewClient creates a new Client instance. If client arg is nil, then http.Client with some
// sensible defaults (e.g., for Timeout) will be used.
func NewClient(cfg *Config, scannedAddr Addr, client *http.Client) (*Client, error) {
	if client == nil {
		client = &http.Client{Timeout: time.Duration(cfg.ClientTimeout) * time.Second}
	}
	return &Client{
		scannedAddr: scannedAddr,
		ralphURL:    cfg.RalphAPIURL,
		apiKey:      cfg.RalphAPIKey,
		client:      client,
	}, nil
}

// NewRequest creates new http.Request object initialized with headers needed for
// communication with Ralph.
func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.apiKey))
	req.Header.Set("User-Agent", "ralph-cli")
	return req, nil
}

// SendToRalph sends to Ralph json-ed datatypes (Ethernet, Memory, etc.) using
// one of the REST methods on a given endpoint. Returned statusCode contains
// the actual HTTP status code, or a special value 0, which designates the case
// when there was an error caused by anything else than HTTP status code > 299.
func (c *Client) SendToRalph(method, endpoint string, data []byte) (statusCode int, err error) {
	url := fmt.Sprintf("%s/%s/", c.ralphURL, endpoint)
	var req *http.Request
	switch {
	case method == "DELETE":
		req, err = c.NewRequest(method, url, nil)
	default:
		req, err = c.NewRequest(method, url, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")
	}
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := readBody(resp)
		if err != nil {
			return 0, err
		}
		err = fmt.Errorf("error while sending to %s with %s method: %s (%s)",
			url, method, body, resp.Status)
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

// GetFromRalph sends a GET request on a given endpoint with specified query.
func (c *Client) GetFromRalph(endpoint string, query string) ([]byte, error) {
	var url string
	switch {
	case query == "":
		url = fmt.Sprintf("%s/%s/", c.ralphURL, endpoint)
	default:
		url = fmt.Sprintf("%s/%s/?%s", c.ralphURL, endpoint, query)
	}
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	resp, err := c.client.Do(req)
	// TODO(xor-xor): When url points to a non-existing Ralph instance, it
	// causes panic somewhere around here.
	defer resp.Body.Close()
	if err != nil {
		return []byte{}, err
	}
	if resp.StatusCode >= 400 {
		return []byte{}, fmt.Errorf("error while sending a GET request to Ralph: %s",
			resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func readBody(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}
	return string(body), nil
}
