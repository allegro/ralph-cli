package main

import "net/http"

// Client provides an interface to interact with Ralph via its REST API.
type Client struct {
	ralphURL   string
	apiKey     string
	apiVersion string
	client     *http.Client
}

// New creates a new Client instance.
func New(ralphURL, apiKey, apiVersion string, client *http.Client) *Client {
	return &Client{
		ralphURL:   ralphURL,
		apiKey:     apiKey,
		apiVersion: apiVersion,
		client:     http.DefaultClient,
	}
}

// Post sends data to Ralph's API via POST method.
func (c *Client) Post(data interface{}) error {
	var err error
	switch data.(type) {
	case PhysicalHost:
		// do something with the data
		err = nil
	case VmHost:
		// do something with the data
		err = nil
	case CloudHost:
		// do something with the data
		err = nil
	case MesosHost:
		// do something with the data
		err = nil
	default:
		err = nil
	}
	return err
}

// Put sends data to Ralph's API via PUT method.
func (c *Client) Put() error {
	return nil
}

// request handles all the low-lowel things (status codes like 40x, 50x, etc.)
// in order to make Post/Put code easier to follow.
func (c *Client) request() error {
	return nil
}
