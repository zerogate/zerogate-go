package zerogate

import (
	"net/http"
)

// Option is a functional option for configuring the API client.
type Option func(*API) error

// HTTPClient accepts a custom *http.Client for making API calls.
func HTTPClient(client *http.Client) Option {
	return func(api *API) error {
		api.httpClient = client
		return nil
	}
}

// BaseURL allows you to override the default HTTP base URL used for API calls.
func BaseURL(baseURL string) Option {
	return func(api *API) error {
		api.BaseURL = baseURL
		return nil
	}
}

// Debug enable debugging
func Debug(debug bool) Option {
	return func(api *API) error {
		api.Debug = debug
		return nil
	}
}

// parseOptions parses the supplied options functions and returns a configured
// *API instance.
func (api *API) parseOptions(options ...Option) error {
	for _, option := range options {
		err := option(api)
		if err != nil {
			return err
		}
	}
	return nil
}
