package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const certificate = "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY----"

// Predefined configurations
var validConfig = APIClientConfig{
	URL:        "http://localhost",
	Hash:       "abcdefghijklmnopqrstuvwxyz",
	Cpus:       4,
	Memory:     1024,
	DeviceType: 1,
}

// Handlers

var registeredHandlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w,
		`{
      "status": "registered",
      "hash": "abcdefghijklmnopqrstuvwxyz",
      "uuid": "00000000-0000-0000-0000-000000000000"
     }`)
}

var unRegisteredHandlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w,
		`{
      "status": "unregistered"
     }`)
}

var waitingClaimHandlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w,
		`{
      "status": "registered"
     }`)
}

var unknownResponseHandlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w,
		`{
      "status": "unknown"
     }`)
}

var wrongJSONHandlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w,
		`{
      "status": "unknown"WRONG
     }D`)
}

var claimedHandlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w,
		`{
      "status": "claimed",
      "cert": "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY----"
     }`)
}

// Helper function to get an ApiClient bounded to a test server
func getTestHTTPClient(handler http.HandlerFunc) (server *httptest.Server, client *http.Client) {
	var transport *http.Transport

	server = httptest.NewServer(handler)

	transport = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client = &http.Client{Transport: transport}

	return server, client
}

////////////////////////////////////////////////////////////////////////////////
/// Start testing
////////////////////////////////////////////////////////////////////////////////

// Test empty configuration structure
func Test_InvalidConfig1(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(APIClientConfig{
		HTTPClient: client,
	})
	defer server.Close()

	assert.Nil(t, apiClient, "apiClient should be nil")
}

// Test invalid configuration structure
func Test_InvalidConfig2(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(APIClientConfig{
		URL:        "http://localhost",
		Cpus:       4,
		Memory:     1024,
		DeviceType: 1,
		HTTPClient: client,
	})
	defer server.Close()

	assert.Nil(t, apiClient, "apiClient should be nil")
}

// Test invalid configuration structure
func Test_InvalidConfig3(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(APIClientConfig{
		URL:        "http://localhost",
		Hash:       "abcdefghijklmnopqrstuvwxyz",
		Memory:     1024,
		DeviceType: 1,
		HTTPClient: client,
	})
	defer server.Close()

	assert.Nil(t, apiClient, "apiClient should be nil")
}

// Test invalid configuration structure
func Test_InvalidConfig4(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(APIClientConfig{
		URL:        "http://localhost",
		Cpus:       4,
		Hash:       "abcdefghijklmnopqrstuvwxyz",
		DeviceType: 1,
		HTTPClient: client,
	})
	defer server.Close()

	assert.Nil(t, apiClient, "apiClient should be nil")
}

// Test invalid configuration structure
func Test_InvalidConfig5(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(APIClientConfig{
		URL:        "http://localhost",
		Cpus:       4,
		Hash:       "abcdefghijklmnopqrstuvwxyz",
		Memory:     1024,
		HTTPClient: client,
	})
	defer server.Close()

	assert.Nil(t, apiClient, "apiClient should be nil")
}

// Test no http client
func Test_InvalidHttpClient(t *testing.T) {
	apiClient := NewAPIClient(validConfig)

	assert.NotNil(t, apiClient, "apiClient should be nil")
}

// Test a valid configuration and http client
func Test_ValidConfig(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client
	defer server.Close()

	assert.NotNil(t, apiClient, "apiClient should be not nil")
}

// Test register success
func Test_Register_Success(t *testing.T) {
	server, client := getTestHTTPClient(registeredHandlerFunc)
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client
	defer server.Close()

	err := apiClient.Register()

	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, "registered", apiClient.status, "Client should be registered")
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", apiClient.uuid, "Wrong UUID")
}

// Test register success
func Test_Register_Fail(t *testing.T) {
	server, client := getTestHTTPClient(unRegisteredHandlerFunc)
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client
	defer server.Close()

	err := apiClient.Register()

	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, "registering", apiClient.status, "Client should be registering")
	assert.Equal(t, "", apiClient.uuid, "Wrong UUID")
}

// Test register fail
func Test_Register_Success_after_fail(t *testing.T) {
	var err error
	server, client := getTestHTTPClient(unRegisteredHandlerFunc)
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client

	err = apiClient.Register()

	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, "registering", apiClient.status, "Client should be registering")
	assert.Equal(t, "", apiClient.uuid, "Wrong UUID")

	server.Close()
	server, client = getTestHTTPClient(registeredHandlerFunc)
	defer server.Close()
	apiClient.config.HTTPClient = client

	err = apiClient.Register()

	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, "registered", apiClient.status, "Client should be registered")
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", apiClient.uuid, "Wrong UUID")
}

// Test verify success when not yet claimed
func Test_Verify_Success_Not_Claimed(t *testing.T) {
	var err error
	server, client := getTestHTTPClient(waitingClaimHandlerFunc)
	defer server.Close()
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client

	apiClient.status = "registered"
	apiClient.uuid = "00000000-0000-0000-0000-000000000000"

	err = apiClient.Verify()

	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, "registered", apiClient.status, "Client should be registered")
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", apiClient.uuid, "Wrong UUID")
}

// Test verify success when claimed
func Test_Verify_Success_Claimed(t *testing.T) {
	var err error
	server, client := getTestHTTPClient(claimedHandlerFunc)
	defer server.Close()
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client

	apiClient.status = "registered"
	apiClient.uuid = "00000000-0000-0000-0000-000000000000"

	err = apiClient.Verify()

	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, "claimed", apiClient.status, "Client should be claimed")
	assert.Equal(t, certificate, apiClient.cert, "Wront certificate")
}

// Test verify unknown response
func Test_Verify_Unknown_Response(t *testing.T) {
	var err error
	server, client := getTestHTTPClient(unknownResponseHandlerFunc)
	defer server.Close()
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client

	apiClient.status = "registered"
	apiClient.uuid = "00000000-0000-0000-0000-000000000000"

	err = apiClient.Verify()

	assert.Error(t, err, "Expected error")
	assert.Equal(t, "registered", apiClient.status, "Client should be claimed")
	assert.Equal(t, "", apiClient.cert, "Wrong certificate")
}

// Test verify unknown response
func Test_Verify_Invalid_JSON(t *testing.T) {
	var err error
	server, client := getTestHTTPClient(wrongJSONHandlerFunc)
	defer server.Close()
	apiClient := NewAPIClient(validConfig)
	apiClient.config.HTTPClient = client

	apiClient.status = "registered"
	apiClient.uuid = "00000000-0000-0000-0000-000000000000"

	err = apiClient.Verify()

	assert.Error(t, err, "Unexpected error")
	assert.Equal(t, "registered", apiClient.status, "Client should be claimed")
	assert.Equal(t, "", apiClient.cert, "Wrong certificate")
}

func Test_IsRegistered(t *testing.T) {
	var result bool
	apiClient := NewAPIClient(validConfig)
	apiClient.status = "registered"

	result = apiClient.IsRegistered()

	assert.Equal(t, true, result, "Should be true")

	apiClient.status = "not registered"

	result = apiClient.IsRegistered()
	assert.Equal(t, false, result, "Should be false")
}

func Test_IsClaimed(t *testing.T) {
	var result bool
	apiClient := NewAPIClient(validConfig)
	apiClient.status = "claimed"

	result = apiClient.IsClaimed()

	assert.Equal(t, true, result, "Should be true")

	apiClient.status = "not claimed"

	result = apiClient.IsClaimed()
	assert.Equal(t, false, result, "Should be false")
}

// Test get certificate success when claimed
func Test_GetCertificate_Success(t *testing.T) {
	apiClient := NewAPIClient(validConfig)

	apiClient.status = "claimed"
	apiClient.cert = certificate

	cert := apiClient.GetCertificate()

	assert.Equal(t, certificate, cert, "Wrong certificate")
}
