package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

type Registerer interface {
	Register()
}
type Verifier interface {
	Verify()
}

// ApiClient is an objet that can communicate with the API to perform a
// registration. It has the necessary methods to interact with the API.
type ApiClient struct {

	// Attributes
	status string         // Current status of the registrtation
	uuid   string         // The UUID that the application should obtain
	cert   string         // Client certificate
	logger *logrus.Logger // Logger

	// Configuration
	config Config
}

// Config stores the client api configuration
type Config struct {
	Url        string // API url
	Hash       string // Required hash to perform the registration
	SleepTime  int    // Time between requests
	Cpus       int    // Number of CPU of the computer
	Memory     int64  // Amount of memory of the computer
	DeviceType int    // Type of the requesting device
	Insecure   bool   // If true, skip SSL verification

	HttpClient *http.Client
	Debug      bool
}

// NewApiClient creates a new instance of an ApiClient object
func NewApiClient(config Config) *ApiClient {
	c := &ApiClient{
		status: "registering",
		logger: logrus.New(),
		config: config,
	}

	if c.config.Debug {
		c.logger.Level = logrus.DebugLevel
	}

	return c
}

// Register send a POST request with some fields to the remote API. It expects
// a UUID from the API.
func (c *ApiClient) Register() error {

	// request structure for register method
	type request struct {
		order      string `json:"order"`
		cpu        int    `json:"cpus"`
		memory     int64  `json:"memory"`
		deviceType int    `json:"type"`
		hash       string `json:"hash"`
	}

	// response structure for register method
	type response struct {
		status string `json:"status"`
		hash   string `json:"hash"`
		uuid   string `json:"uuid"`
	}

	// Build the request
	req := request{
		order:      "register",
		cpu:        c.config.Cpus,
		memory:     c.config.Memory,
		deviceType: c.config.DeviceType,
		hash:       c.config.Hash,
	}

	// Generate a JSON message with the request
	marshalledReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Send the JSON request until the registration procedure succeeds
	for {
		c.logger.Debugf("Sending register request")
		bufferReq := bytes.NewBuffer(marshalledReq)
		rawResponse, err := http.Post(c.config.Url, "application/json", bufferReq)
		if err != nil {
			return err
		}
		defer rawResponse.Body.Close()

		// Read response to a buffer
		var bufferResponse []byte
		rawResponse.Body.Read(bufferResponse)

		// Unmarshall the response
		res := response{}
		err = json.Unmarshal(bufferResponse, &res)
		if err != nil {
			return err
		}

		// Check response
		if res.status == "registered" {
			c.uuid = res.uuid
			c.status = res.status
			c.logger.Debugf("Got UUID: %s", c.uuid)
			return nil
		}

		// Wait before the next request
		time.Sleep(time.Duration(c.config.SleepTime) * time.Millisecond)
	}
}

// Verify send the UUID along with the HASH to the API and expect to receive
// a client certificate
func (c *ApiClient) Verify() error {

	// request structure for register method
	type request struct {
		order string `json:"order"`
		hash  string `json:"hash"`
		uuid  string `json:"uuid"`
	}

	// response structure for register method
	type response struct {
		status string `json:"status"`
		cert   string `json:"cert"`
	}

	// Build the request
	req := request{
		order: "verify",
		hash:  c.config.Hash,
		uuid:  c.uuid,
	}

	// Generate a JSON message with the request
	marshalledReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Send the JSON request until the registration procedure succeeds
	for {
		c.logger.Debugf("Sending verify request")
		bufferReq := bytes.NewBuffer(marshalledReq)
		rawResponse, err := http.Post(c.config.Url, "application/json", bufferReq)
		if err != nil {
			return err
		}
		defer rawResponse.Body.Close()

		// Read response to a buffer
		var bufferResponse []byte
		rawResponse.Body.Read(bufferResponse)

		// Unmarshall the response
		res := response{}
		err = json.Unmarshal(bufferResponse, &res)
		if err != nil {
			return err
		}

		// Check response
		switch res.status {
		case "registered":
			c.logger.Debugf("Waiting to be claimed")
			break
		case "claimed":
			c.cert = res.cert
			c.status = res.status
			c.logger.Debugf("Got certificate")
			return nil
			break
		default:
			c.logger.Warnf("Unknow response status: %s", res.status)
		}

		// Wait before the next request
		time.Sleep(time.Duration(c.config.SleepTime) * time.Millisecond)
	}
}

// Return the certificate if the device is claimed
func (c *ApiClient) GetCertificate() (cert string, err error) {
	if c.status == "claimed" {
		cert = c.cert
		return
	} else {
		err = errors.New("Device is not yet claimed")
		return
	}
}
