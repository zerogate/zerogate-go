package zerogate

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"
	"time"
)

type service struct {
	client *Client
}

// Client holds the configuration for the current API client.
type Client struct {
	mutex      sync.RWMutex
	apiKey     string
	apiSecret  string
	baseUrl    string
	debug      bool
	userAgent  string
	headers    http.Header
	httpClient *http.Client
	logger     *log.Logger

	common service

	Tenant *TenantService
}

// newClient provides shared logic for New.
func newClient(opts ...Option) (*Client, error) {
	silentLogger := log.New(io.Discard, "", log.LstdFlags)

	client := &Client{
		baseUrl:   baseUrl,
		userAgent: userAgent,
		headers:   make(http.Header),
		logger:    silentLogger,
	}
	client.common.client = client

	err := client.parseOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("options parsing failed: %w", err)
	}

	if client.httpClient == nil {
		client.httpClient = http.DefaultClient
	}

	client.Tenant = (*TenantService)(&client.common)

	return client, nil
}

// New creates a new ZeroGate API client.
func New(key, secret string, opts ...Option) (*Client, error) {
	if key == "" || secret == "" {
		return nil, errors.New(errEmptyCredentials)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.apiKey = key
	api.apiSecret = secret

	return api, nil
}

func (c *Client) getClient() *http.Client {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	httpClient := c.httpClient
	clientCopy := *httpClient
	return &clientCopy
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, query map[string][]string, body interface{}, headers http.Header) (*APIResponse, error) {
	var err error
	var resp *http.Response
	var respBody []byte

	c.mutex.RLock()
	apiKey := c.apiKey
	apiSecret := c.apiSecret
	baseUrl := c.baseUrl
	debug := c.debug
	userAgent := c.userAgent
	apiHeaders := c.headers
	c.mutex.RUnlock()

	var reqBody io.Reader
	if body != nil && (method == http.MethodPost || method == http.MethodPut) {
		if r, ok := body.(io.Reader); ok {
			reqBody = r
		} else if bodyBytes, ok := body.([]byte); ok {
			reqBody = bytes.NewReader(bodyBytes)
		} else {
			var jsonBody []byte
			jsonBody, err = json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("error marshalling body to JSON: %w", err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}
	} else if method == http.MethodPost || method == http.MethodPut {
		reqBody = bytes.NewReader([]byte("{}"))
	}
	var bodyBytes []byte

	if method == http.MethodPost || method == http.MethodPut {
		bodyBytes, err = io.ReadAll(reqBody)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
		reqBody = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Get the current datetime in ISO 8601 format
	now := time.Now().Unix()

	req, err := http.NewRequestWithContext(ctx, method, baseUrl+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("ZeroGate request creation failed: %w", err)
	}
	// Convert the map to a URL query string
	values := url.Values{}
	for k, v := range query {
		for _, vv := range v {
			values.Add(k, vv)
		}
	}
	queryString := values.Encode()
	req.URL.RawQuery = queryString

	combinedHeaders := make(http.Header)
	for k, v := range apiHeaders {
		combinedHeaders[k] = v
	}
	for k, v := range headers {
		combinedHeaders[k] = v
	}
	req.Header = combinedHeaders

	// Combine the method, endpoint, and datetime into the message to sign
	message := req.Method + req.URL.Path + fmt.Sprint(now)

	// Create an HMAC-SHA512 hash using the API secret as the key
	h := hmac.New(sha512.New, []byte(apiSecret))
	h.Write([]byte(message))
	if method == http.MethodPost || method == http.MethodPut {
		h.Write(bodyBytes)
	}
	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("Authorization", fmt.Sprintf("APIKey=%s, Signature=%s, Nonce=%d", apiKey, signature, now))
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		// strip out any sensitive information from the request payload.
		sensitiveKeys := []string{apiKey, apiSecret}
		for _, key := range sensitiveKeys {
			if key != "" {
				valueRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", key))
				dump = valueRegex.ReplaceAll(dump, []byte("[**************]"))
			}
		}
		log.Printf("\n%s", string(dump))
	}
	client := c.getClient()
	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ZeroGate request failed: %w", err)
	}
	defer resp.Body.Close()
	if debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		log.Printf("\n%s", string(dump))
	}
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response read failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var r ErrorResponse
		err = json.Unmarshal(respBody, &r)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		err = &Error{
			StatusCode: resp.StatusCode,
			Response:   r,
		}
		return nil, err
	}

	return &APIResponse{
		Body:       respBody,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
	}, nil
}

func (c *Client) get(ctx context.Context, endpoint string, query map[string][]string, headers http.Header) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodGet, endpoint, query, nil, headers)
}

func (c *Client) post(ctx context.Context, endpoint string, query map[string][]string, body interface{}, headers http.Header) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, endpoint, query, body, headers)
}

func (c *Client) put(ctx context.Context, endpoint string, query map[string][]string, body interface{}, headers http.Header) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPut, endpoint, query, body, headers)
}

func (c *Client) delete(ctx context.Context, endpoint string, query map[string][]string, headers http.Header) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, endpoint, query, nil, headers)
}
