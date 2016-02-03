package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
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
	Memory     uint64 // Amount of memory of the computer
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
	type Request struct {
		Order      string `json:"order"`
		Cpu        int    `json:"cpus"`
		Memory     uint64 `json:"memory"`
		DeviceType int    `json:"type"`
		Hash       string `json:"hash"`
	}

	// response structure for register method
	type Response struct {
		Status string `json:"status"`
		Hash   string `json:"hash"`
		Uuid   string `json:"uuid"`
	}

	// Build the request
	req := Request{
		Order:      "register",
		Cpu:        c.config.Cpus,
		Memory:     c.config.Memory,
		DeviceType: c.config.DeviceType,
		Hash:       c.config.Hash,
	}

	// Generate a JSON message with the request
	marshalledReq, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	// Send request
	log.Debugf("Sending register request")
	bufferReq := bytes.NewBuffer(marshalledReq)
	httpReq, err := http.NewRequest("POST", c.config.Url, bufferReq)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawResponse, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer rawResponse.Body.Close()

	// Read response to a buffer
	bufferResponse, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		return err
	}

	// Unmarshall the response
	res := Response{}
	err = json.Unmarshal(bufferResponse, &res)
	if err != nil {
		return err
	}

	// Check response
	if res.Status == "registered" {
		c.uuid = res.Uuid
		c.status = res.Status
		log.Debugf("Got UUID: %s", c.uuid)
	}

	return nil
}

// Verify send the UUID along with the HASH to the API and expect to receive
// a client certificate
func (c *ApiClient) Verify() error {

	// request structure for register method
	type request struct {
		Order string `json:"order"`
		Hash  string `json:"hash"`
		Uuid  string `json:"uuid"`
	}

	// response structure for register method
	type response struct {
		Status string `json:"status"`
		Cert   string `json:"cert"`
	}

	// Build the request
	req := request{
		Order: "verify",
		Hash:  c.config.Hash,
		Uuid:  c.uuid,
	}

	// Generate a JSON message with the request
	marshalledReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Send request
	log.Debugf("Sending register request")
	bufferReq := bytes.NewBuffer(marshalledReq)
	httpReq, err := http.NewRequest("POST", c.config.Url, bufferReq)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawResponse, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer rawResponse.Body.Close()

	// Read response to a buffer
	bufferResponse, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		return err
	}

	// Unmarshall the response
	res := response{}
	err = json.Unmarshal(bufferResponse, &res)
	if err != nil {
		return err
	}

	// Check response
	switch res.Status {
	case "registered":
		log.Debugf("Waiting to be claimed")
		break
	case "claimed":
		c.cert = res.Cert
		c.status = res.Status
		log.Debugf("Got certificate")
		break
	default:
		log.Warnf("Unknow response status: %s", res.Status)
		break
	}

	return nil
}

func (c *ApiClient) IsRegistered() bool {
	if c.status == "registered" {
		return true
	} else {
		return false
	}
}

func (c *ApiClient) IsClaimed() bool {
	if c.status == "claimed" {
		return true
	} else {
		return false
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
