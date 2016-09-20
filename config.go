package main

import (
	"database/sql"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// APIClientConfig stores the client api configuration
type APIClientConfig struct {
	Insecure   bool           // If true, skip SSL verification
	URL        string         // API url
	Hash       string         // Required hash to perform the registration
	Cpus       int            // Number of CPU of the computer
	Memory     uint64         // Amount of memory of the computer
	DeviceType int            // Type of the requesting device
	Logger     *logrus.Logger // Logger to use
	HTTPClient *http.Client   // HTTP Client to wrap
}

// DatabaseConfig stores the database configuration
type DatabaseConfig struct {
	sqldb  *sql.DB
	dbFile string
	Logger *logrus.Logger // Logger to use
}
