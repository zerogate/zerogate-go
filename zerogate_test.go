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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	// client is the API client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server

	// router is the HTTP request router used with the test server.
	router *gin.Engine
)

const (
	testApiKey    = "key_5fbea6690113a5b9560bc9def29c91e2"
	testApiSecret = "1f4f6db557e4fdce6eb1dbbcc9f5d544f99252e8c2b5158a566e1c4667a48717"
)

func setup(opts ...Option) {
	// test server
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	router.Use(gin.Recovery())
	server = httptest.NewServer(router)

	// ZeroGate client configured to use test server
	client, _ = New(testApiKey, testApiSecret, opts...)
	client.baseUrl = server.URL
}

func teardown() {
	server.Close()
}

func TestClient_Headers(t *testing.T) {
	// it should set default headers
	setup()
	router.GET("/get", func(c *gin.Context) {
		assert.Equal(t, http.MethodGet, c.Request.Method, "Expected method 'GET', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		c.JSON(http.StatusOK, "ok")
	})
	router.POST("/post", func(c *gin.Context) {
		assert.Equal(t, http.MethodPost, c.Request.Method, "Expected method 'POST', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		c.JSON(http.StatusOK, "ok")
	})
	router.PUT("/put", func(c *gin.Context) {
		assert.Equal(t, http.MethodPut, c.Request.Method, "Expected method 'PUT', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		c.JSON(http.StatusOK, "ok")
	})
	router.DELETE("/delete", func(c *gin.Context) {
		assert.Equal(t, http.MethodDelete, c.Request.Method, "Expected method 'DELETE', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		c.JSON(http.StatusOK, "ok")
	})
	header := make(http.Header)
	client.doRequest(context.Background(), http.MethodGet, "/get", nil, nil, header)
	client.doRequest(context.Background(), http.MethodPost, "/post", nil, nil, header)
	client.doRequest(context.Background(), http.MethodPut, "/put", nil, nil, header)
	client.doRequest(context.Background(), http.MethodDelete, "/delete", nil, nil, header)
	teardown()
}

func testSignature(c *gin.Context, t *testing.T) {
	// Get the authorization header
	authHeader := c.Request.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader, "empty Authorization header")

	// Split the authorization header into its components
	authParts := strings.Split(authHeader, ", ")
	if len(authParts) != 3 {
		assert.NoError(t, nil, "invalid Authorization header")
		return
	}

	// Parse the authorization header for the API key, signature, and nonce
	var apiKey, signature string
	var nonce int64

	if n, err := fmt.Sscanf(authParts[0], "APIKey=%s", &apiKey); err != nil || n != 1 {
		assert.NoError(t, err, "invalid Authorization header")
		return
	}
	if n, err := fmt.Sscanf(authParts[1], "Signature=%s", &signature); err != nil || n != 1 {
		assert.NoError(t, err, "invalid Authorization header")
		return
	}
	if n, err := fmt.Sscanf(authParts[2], "Nonce=%d", &nonce); err != nil || n != 1 {
		assert.NoError(t, err, "invalid Authorization header")
		return
	}

	// Combine the HTTP method, endpoint, nonce, and request body into the message to sign
	method := c.Request.Method
	endpoint := c.Request.URL.Path
	message := method + endpoint + fmt.Sprint(nonce)

	// Create an HMAC-SHA512 hash using the API secret as the key
	h := hmac.New(sha512.New, []byte(testApiSecret))
	h.Write([]byte(message))
	if method == http.MethodPost || method == http.MethodPut {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			assert.NoError(t, err, "error reading body")
			return
		}
		c.Request.Body.Close() //  must close
		h.Write(bodyBytes)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	reqSignature := hex.EncodeToString(h.Sum(nil))

	assert.Equalf(t, signature, reqSignature, "signature mismatch expected %s got %s", signature, reqSignature)
}

func TestContextTimeout(t *testing.T) {
	setup()
	defer teardown()

	handler := func(c *gin.Context) {
		time.Sleep(3 * time.Second)
		c.JSON(http.StatusOK, "ok")
	}
	router.GET("/timeout", handler)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	_, err := client.doRequest(ctx, http.MethodGet, "/timeout", nil, nil, nil)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.WithinDuration(t, start, time.Now(), 2*time.Second,
		"doRequest took too much time with an expiring context")
}
