package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// APIEndpoints maps datatype names to Ralph's API endpoints associated with them.
var APIEndpoints = map[string]string{
	"PhysicalHost":      "data-center-assets",
	"VMHost":            "virtual-servers",
	"CloudHost":         "cloud-hosts",
	"EthernetComponent": "ethernets",
	"BaseObject":        "base-objects",
	"IPAddress":         "ipaddresses", // only for ExcludeMgmt's purposes!
	// ...and so on for other data types defined for Ralph
}

// Client provides an interface to interact with Ralph via its REST API.
type Client struct {
	scannedAddr Addr
	ralphURL    string
	apiKey      string
	apiVersion  string // Not used b/c Ralph doesn't have any API versioning (yet).
	client      *http.Client
}

// NewClient creates a new Client instance.
func NewClient(ralphURL, apiKey string, scannedAddr Addr, client *http.Client) (*Client, error) {
	// TODO(xor-xor): get rid of Query/Fragment if present
	if apiKey == "" {
		return nil, fmt.Errorf("API key is missing (did you forget to set it via RALPH_API_KEY environment variable?)")
	}
	if ralphURL == "" {
		return nil, fmt.Errorf("Ralph's URL is missing (did you forget to set it via RALPH_API_URL environment variable?)")
	}
	// TODO(xor-xor): Investigate why url.Parse happily accepts stuff like "httplocalhost" or
	// "http/localhost/api", and add some additional checks here for such cases.
	u, err := url.Parse(ralphURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing Ralph's URL: %v", err)
	}
	if client == nil {
		// TODO(xor-xor): This timeout should be taken from config.
		client = &http.Client{Timeout: time.Second * 10}
	}
	return &Client{
		scannedAddr: scannedAddr,
		ralphURL:    u.String(),
		apiKey:      apiKey,
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

// SendToRalph sends to Ralph json-ed datatypes (EthernetComponent, PhysicalHost, etc.)
// using one of the REST methods on a given endpoint.
// Returned statusCode contains the actual HTTP status code, or a special value 0, which designates
// the case when there was an error caused by anything else than HTTP status code > 299.
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
	if resp.StatusCode > 299 {
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
	url := fmt.Sprintf("%s/%s/?%s", c.ralphURL, endpoint, query)
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, nil
}

func readBody(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}
	return string(body), nil
}
