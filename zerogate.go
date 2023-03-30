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
	"time"
)

// API holds the configuration for the current API client.
type API struct {
	APIKey     string
	APISecret  string
	BaseURL    string
	Debug      bool
	UserAgent  string
	headers    http.Header
	httpClient *http.Client
	logger     *log.Logger
}

// newClient provides shared logic for New.
func newClient(opts ...Option) (*API, error) {
	silentLogger := log.New(io.Discard, "", log.LstdFlags)

	api := &API{
		BaseURL:   baseUrl,
		UserAgent: userAgent,
		headers:   make(http.Header),
		logger:    silentLogger,
	}

	err := api.parseOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("options parsing failed: %w", err)
	}

	if api.httpClient == nil {
		api.httpClient = http.DefaultClient
	}

	return api, nil
}

// New creates a new ZeroGate API client.
func New(key, secret string, opts ...Option) (*API, error) {
	if key == "" || secret == "" {
		return nil, errors.New(errEmptyCredentials)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIKey = key
	api.APISecret = secret

	return api, nil
}

func (api *API) doRequest(ctx context.Context, method, endpoint string, query map[string][]string, body interface{}, headers http.Header) (*APIResponse, error) {
	var err error
	var resp *http.Response
	var respBody []byte

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

	// Get the current datetime in ISO 8601 format
	now := time.Now().Unix()

	req, err := http.NewRequestWithContext(ctx, method, api.BaseURL+endpoint, reqBody)
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
	for k, v := range api.headers {
		combinedHeaders[k] = v
	}
	for k, v := range headers {
		combinedHeaders[k] = v
	}
	req.Header = combinedHeaders

	// Combine the method, endpoint, and datetime into the message to sign
	message := req.Method + req.URL.Path + fmt.Sprint(now)

	// Create an HMAC-SHA512 hash using the API secret as the key
	h := hmac.New(sha512.New, []byte(api.APISecret))
	h.Write([]byte(message))
	if method == http.MethodPost || method == http.MethodPut {
		bodyBytes, err := io.ReadAll(reqBody)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
		h.Write(bodyBytes)
		reqBody = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("Authorization", fmt.Sprintf("APIKey=%s, Signature=%s, Nonce=%d", api.APIKey, signature, now))
	if api.UserAgent != "" {
		req.Header.Set("User-Agent", api.UserAgent)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if api.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		// strip out any sensitive information from the request payload.
		sensitiveKeys := []string{api.APIKey, api.APISecret}
		for _, key := range sensitiveKeys {
			if key != "" {
				valueRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", key))
				dump = valueRegex.ReplaceAll(dump, []byte("[**************]"))
			}
		}
		log.Printf("\n%s", string(dump))
	}

	resp, err = api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ZeroGate request failed: %w", err)
	}
	defer resp.Body.Close()

	if api.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		log.Printf("\n%s", string(dump))
	}
	respBody, err = io.ReadAll(resp.Body)

	return &APIResponse{
		Body:       respBody,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
	}, nil
}
