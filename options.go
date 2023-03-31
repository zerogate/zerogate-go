package zerogate

import (
	"net/http"
)

// Option is a functional option for configuring the API client.
type Option func(*Client) error

// HTTPClient accepts a custom *http.Client for making API calls.
func HTTPClient(httpClient *http.Client) Option {
	return func(client *Client) error {
		client.httpClient = httpClient
		return nil
	}
}

// BaseURL allows you to override the default HTTP base URL used for API calls.
func BaseURL(baseURL string) Option {
	return func(client *Client) error {
		client.baseUrl = baseURL
		return nil
	}
}

// Debug enable debugging
func Debug(debug bool) Option {
	return func(client *Client) error {
		client.debug = debug
		return nil
	}
}

// parseOptions parses the supplied options functions and returns a configured
// *Client instance.
func (c *Client) parseOptions(options ...Option) error {
	for _, option := range options {
		err := option(c)
		if err != nil {
			return err
		}
	}
	return nil
}
