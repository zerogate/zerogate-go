package zerogate

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the API client being tested.
	client *API

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

const (
	testApiKey    = "key_5fbea6690113a5b9560bc9def29c91e2"
	testApiSecret = "1f4f6db557e4fdce6eb1dbbcc9f5d544f99252e8c2b5158a566e1c4667a48717"
)

func setup(opts ...Option) {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// ZeroGate client configured to use test server
	client, _ = New(testApiKey, testApiSecret, opts...)
	client.BaseURL = server.URL
}

func teardown() {
	server.Close()
}

func TestClient_Headers(t *testing.T) {
	// it should set default headers
	setup()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		testSignature(r, t)
	})
	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		testSignature(r, t)
	})
	mux.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		testSignature(r, t)
	})
	mux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		testSignature(r, t)
	})
	header := make(http.Header)
	client.doRequest(context.Background(), http.MethodGet, "/get", nil, nil, header)
	client.doRequest(context.Background(), http.MethodPost, "/post", nil, nil, header)
	client.doRequest(context.Background(), http.MethodPut, "/put", nil, nil, header)
	client.doRequest(context.Background(), http.MethodDelete, "/delete", nil, nil, header)
	teardown()
}

func testSignature(r *http.Request, t *testing.T) {
	// Get the authorization header
	authHeader := r.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader, "empty Authorization header")

	// Split the authorization header into its components
	authParts := strings.Split(authHeader, ", ")
	if len(authParts) != 3 {
		assert.Error(t, nil, "invalid Authorization header")
	}

	// Parse the authorization header for the API key, signature, and nonce
	var apiKey, signature string
	var nonce int64

	if n, err := fmt.Sscanf(authParts[0], "APIKey=%s", &apiKey); err != nil || n != 1 {
		assert.Error(t, err, "invalid Authorization header")
	}
	if n, err := fmt.Sscanf(authParts[1], "Signature=%s", &signature); err != nil || n != 1 {
		assert.Error(t, err, "invalid Authorization header")
	}
	if n, err := fmt.Sscanf(authParts[2], "Nonce=%d", &nonce); err != nil || n != 1 {
		assert.Error(t, err, "invalid Authorization header")
	}

	// Combine the HTTP method, endpoint, nonce, and request body into the message to sign
	method := r.Method
	endpoint := r.URL.Path
	message := method + endpoint + fmt.Sprint(nonce)

	// Create an HMAC-SHA512 hash using the API secret as the key
	h := hmac.New(sha512.New, []byte(testApiSecret))
	h.Write([]byte(message))
	if method == http.MethodPost || method == http.MethodPut {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			assert.Error(t, err, "error reading body")
		}
		r.Body.Close() //  must close
		h.Write(bodyBytes)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	reqSignature := hex.EncodeToString(h.Sum(nil))

	assert.Equalf(t, signature, reqSignature, "signature mismatch expected %s got %s", signature, reqSignature)
}

func TestContextTimeout(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}
	mux.HandleFunc("/timeout", handler)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	_, err := client.doRequest(ctx, http.MethodGet, "/timeout", nil, nil, nil)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.WithinDuration(t, start, time.Now(), 2*time.Second,
		"doRequest took too much time with an expiring context")
}
