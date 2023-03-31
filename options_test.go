package zerogate

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestDebugOption(t *testing.T) {
	client, err := New(testApiKey, testApiSecret, Debug(true))
	if err != nil {
		assert.Error(t, nil, "client creation failed")
	}
	assert.True(t, client.debug, "client debug should be enabled")
	client, err = New(testApiKey, testApiSecret, Debug(false))
	if err != nil {
		assert.Error(t, nil, "client creation failed")
	}
	assert.False(t, client.debug, "client debug should not be enabled")
}

func TestHttpClientOption(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	client, err := New(testApiKey, testApiSecret, HTTPClient(httpClient))
	if err != nil {
		assert.Error(t, nil, "client creation failed")
	}
	assert.Equal(t, client.httpClient, httpClient, "HTTPClient is not equal")
	assert.Equal(t, client.httpClient.Timeout, httpClient.Timeout, "timeout is not equal")
}

func TestBaseUrlOption(t *testing.T) {
	testBaseUrl := "http://locahost:8080/public/v1"
	client, err := New(testApiKey, testApiSecret, BaseURL(testBaseUrl))
	if err != nil {
		assert.Error(t, nil, "client creation failed")
	}
	assert.Equal(t, client.baseUrl, testBaseUrl, "base url is not equal")
}
