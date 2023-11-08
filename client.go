// Copyright (C) 2016 Eneo Tecnologia S.L.
// Diego Fernández Barrera <bigomby@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	registerRequest    = "register"
	verifyRequest      = "verify"
	claimedResponse    = "claimed"
	registeredResponse = "registered"
)

// APIClient is an objet that can communicate with the API to perform a
// registration. It has the necessary methods to interact with the API.
type APIClient struct {
	status   string // Current status of the registrtation
	cert     string // Client certificate
	nodename string // Name of the node received along with the cert

	config APIClientConfig
}

// NewAPIClient creates a new instance of an ApiClient object
func NewAPIClient(config APIClientConfig) *APIClient {
	c := &APIClient{
		config: config,
		status: "registering",
	}

	if c.config.Logger == nil {
		c.config.Logger = logrus.NewEntry(logrus.New())
		c.config.Logger.Logger.Out = ioutil.Discard
	} else {
		c.config.Logger = c.config.Logger.WithFields(logrus.Fields{
			"component": "api_client",
		})
	}

	logger := c.config.Logger

	// Check if the configuration is ok
	if len(c.config.URL) == 0 {
		logger.Warnf("Url not provided")
		return nil
	}
	if len(c.config.Hash) == 0 {
		logger.Warnf("Hash not provided")
		return nil
	}
	if c.config.Cpus == 0 {
		logger.Warnf("CPU number not provided")
		return nil
	}
	if c.config.Memory == 0 {
		logger.Warnf("Memory not provided")
		return nil
	}
	if c.config.DeviceType == 0 {
		logger.Warnf("Device type not provided")
		return nil
	}
	if c.config.HTTPClient == nil {
		if c.config.Insecure {
			c.config.HTTPClient = &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}
		} else {
			c.config.HTTPClient = &http.Client{}
		}
	}

	return c
}

// Register send a POST request with some fields to the remote API. It expects
// a UUID from the API.
func (c *APIClient) Register() (uuid string, err error) {
	logger := c.config.Logger

	if c.status == registeredResponse {
		return "", errors.New("This device is already registered")
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
		UUID   string `json:"uuid"`
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
		return "", err
	}

	// Send request
	logger.Debugf("Register request: %v", req)
	bufferReq := bytes.NewBuffer(marshalledReq)
	httpReq, err := http.NewRequest("POST", c.config.URL, bufferReq)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawResponse, err := c.config.HTTPClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer rawResponse.Body.Close()
	if rawResponse.StatusCode >= 400 {
		return "", errors.New("Got status code: " + rawResponse.Status)
	}

	// Read response to a buffer
	bufferResponse, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		return "", err
	}

	// Unmarshall the response
	res := Response{}
	err = json.Unmarshal(bufferResponse, &res)
	if err != nil {
		return "", err
	}

	logger.Debugf("Register response: %v", res)

	// Check response
	if res.Status == registeredResponse {
		c.status = res.Status
	}

	return res.UUID, nil
}

// Verify send the UUID along with the HASH to the API and expect to receive
// a client certificate
func (c *APIClient) Verify(uuid string) error {
	logger := c.config.Logger

	if c.status == claimedResponse {
		return errors.New("This device is already claimed")
	}

	// request structure for register method
	type request struct {
		Order string `json:"order"`
		Hash  string `json:"hash"`
		UUID  string `json:"uuid"`
	}

	// response structure for register method
	type response struct {
		Status   string `json:"status"`
		Cert     string `json:"cert"`
		Nodename string `json:"nodename"`
	}

	// Build the request
	req := request{
		Order: "verify",
		Hash:  c.config.Hash,
		UUID:  uuid,
	}

	// Generate a JSON message with the request
	marshalledReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Send request
	logger.Debugf("Verify request: %v", req)
	bufferReq := bytes.NewBuffer(marshalledReq)
	httpReq, err := http.NewRequest("POST", c.config.URL, bufferReq)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawResponse, err := c.config.HTTPClient.Do(httpReq)
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
	res := response{}
	err = json.Unmarshal(bufferResponse, &res)
	if err != nil {
		return err
	}

	logger.Debugf("Claimed response: %v", res)

	if res.Status == registeredResponse {
		return nil
	}

	if res.Status == claimedResponse {
		c.nodename = res.Nodename
		c.status = res.Status
		c.cert = res.Cert

		return nil
	}

	return errors.New("Unknow status: " + res.Status)
}

// IsRegistered check if the client has been registered previously
func (c *APIClient) IsRegistered() bool {
	return c.status == registeredResponse
}

// IsClaimed check if the client has been claimed previously
func (c *APIClient) IsClaimed() bool {
	return c.status == claimedResponse
}

// GetCertificate returns the certificate if the device is claimed
func (c *APIClient) GetCertificate() string {
	return c.cert
}

// GetNodename return the certificate if the device is claimed
func (c *APIClient) GetNodename() string {
	return c.nodename
}
