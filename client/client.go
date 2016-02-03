package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"redborder/rb-register-2/database"
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
	status     string // Current status of the registrtation
	uuid       string // The UUID that the application should obtain
	cert       string // Client certificate
	HttpClient *http.Client
	db         *database.Database

	// Configuration
	config Config
}

// Config stores the client api configuration
type Config struct {
	Url        string // API url
	Hash       string // Required hash to perform the registration
	Cpus       int    // Number of CPU of the computer
	Memory     uint64 // Amount of memory of the computer
	DeviceType int    // Type of the requesting device
	Debug      bool   // Show debug info
}

var log *logrus.Logger

// NewApiClient creates a new instance of an ApiClient object
func NewApiClient(config Config, httpClient *http.Client, db *database.Database) *ApiClient {
	log = logrus.New()

	// Show debuf info
	if config.Debug {
		log.Level = logrus.DebugLevel
	}

	// Check if the configuration is ok
	if len(config.Url) == 0 {
		log.Errorf("Url not provided")
		return nil
	}
	if len(config.Hash) == 0 {
		log.Errorf("Hash not provided")
		return nil
	}
	if config.Cpus == 0 {
		log.Errorf("CPU number not provided")
		return nil
	}
	if config.Memory == 0 {
		log.Errorf("Memory not provided")
		return nil
	}
	if config.DeviceType == 0 {
		log.Errorf("Device type not provided")
		return nil
	}
	if httpClient == nil {
		log.Errorf("Invalid HTTP client")
		return nil
	}

	// Create instance of ApiClient
	c := &ApiClient{
		status:     "registering",
		HttpClient: httpClient,
		config:     config,
	}

	// If a DB has been provided try to find the hash on it
	if db != nil {
		c.db = db
		uuid, err := c.db.LoadUuid(config.Hash)
		if err != nil {
			log.Error("Error loading UUID from DB")
			return nil
		}

		// If the provided hash is already on database, load the associated UUID
		// and set status to registered
		if len(uuid) > 0 {
			c.uuid = uuid
			c.status = "registered"
		}
	}

	return c
}

// Register send a POST request with some fields to the remote API. It expects
// a UUID from the API.
func (c *ApiClient) Register() error {

	if c.status == "registered" {
		return errors.New("This device is already registered")
	}

	// request structure for register method
	type Request struct {
		Order      string `json:"order"`
		Cpus       int    `json:"cpus"`
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
		Cpus:       c.config.Cpus,
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
	if rawResponse.StatusCode >= 400 {
		return errors.New("Got status code: " + rawResponse.Status)
	}

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
		if c.db != nil {
			c.db.StoreUuid(c.config.Hash, res.Uuid)
		}
	}

	return nil
}

// Verify send the UUID along with the HASH to the API and expect to receive
// a client certificate
func (c *ApiClient) Verify() error {

	if c.status == "claimed" {
		return errors.New("This device is already claimed")
	}

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
	log.Debugf("Sending verify request")
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
		log.Warnf("Unknow response: %s", res)
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
