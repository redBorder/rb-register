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
	"flag"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/capnm/sysinfo"
	_ "github.com/mattn/go-sqlite3"
)

var version string
var goVersion = runtime.Version()

var (
	debug         *bool       // Debug flag
	apiURL        *string     // API url
	hash          *string     // Required hash to perform the registration
	deviceAlias   *string     // Given alias of the device
	sleepTime     *int        // Time between requests
	insecure      *bool       // If true, skip SSL verification
	certFile      *string     // Path to store de certificate
	dbFile        *string     // File to persist the state
	daemonFlag    *bool       // Start in daemon mode
	pid           *string     // Path to PID file
	logFile       *string     // Log file
	nodenameFile  *string     // File to store nodename
	scriptFile    *string     // Script to call after the certificate has been obtained
	scriptLogFile *string     // Log to save the result of the script called
	si            *sysinfo.SI // System information
)

// Global logger
var logger = logrus.New()

// init parses flags
func init() {
	scriptFile = flag.String("script", "/opt/rb/bin/rb_register_finish.sh", "Script to call after the certificate has been obtained")
	debug = flag.Bool("debug", false, "Show debug info")
	apiURL = flag.String("url", "http://localhost", "Protocol and hostname to connect")
	hash = flag.String("hash", "00000000-0000-0000-0000-000000000000", "Hash to use in the request")
	sleepTime = flag.Int("sleep", 300, "Time between requests in seconds")
	deviceAlias = flag.String("type", "", "Type of the registering device")
	insecure = flag.Bool("no-check-certificate", false, "Dont check if the certificate is valid")
	certFile = flag.String("cert", "/opt/rb/etc/chef/client.pem", "Certificate file")
	dbFile = flag.String("db", "", "File to persist the state")
	daemonFlag = flag.Bool("daemon", false, "Start in daemon mode")
	pid = flag.String("pid", "pid", "File containing PID")
	logFile = flag.String("log", "log", "Log file")
	nodenameFile = flag.String("nodename", "", "File to store nodename")
	versionFlag := flag.Bool("version", false, "Display version")

	flag.Parse()

	if *versionFlag {
		displayVersion()
		os.Exit(0)
	}

	// Init logger
	if *debug {
		logger.Level = logrus.DebugLevel
	}
}

func main() {
	var db *Database
	if len(*deviceAlias) == 0 {
		flag.Usage()
		logger.Fatal("You must provide a device alias")
	}

	deviceType, err := getDeviceType(*deviceAlias)
	if err != nil {
		logger.Fatal("Invalid device alias")
	}

	si = sysinfo.Get()

	if *daemonFlag {
		daemonize()
	}

	if len(*dbFile) > 0 {
		db = NewDatabase(DatabaseConfig{dbFile: *dbFile})
		if db == nil {
			logger.Errorln("Error opening database")
			halt()
		}
		defer db.Close()
	}

	// Create a new API client for handle the connection with the API
	apiClient := NewAPIClient(
		APIClientConfig{
			URL:        *apiURL,
			Hash:       *hash,
			Cpus:       runtime.NumCPU(),
			Memory:     si.TotalRam,
			DeviceType: deviceType,
			Insecure:   *insecure,
		},
	)

	uuid, err := registrationProcess(apiClient, db)
	if err != nil {
		logger.Errorf("Registration failed: %v", err)
		halt()
	}
	logger.Infoln("Registration completed")

	cert, nodename, err := verificationProcess(uuid, apiClient, db)
	if err != nil {
		logger.Errorf("Verification failed: %v", err)
		halt()
	}
	logger.Infoln("Verification completed")

	if len(cert) > 0 && certFile != nil {
		if err := ioutil.WriteFile(*certFile, []byte(cert), os.ModePerm); err != nil {
			logger.Fatalf("Error saving certificate: %s", err.Error())
		} else {
			logger.Debugf("Certificate saved on %s", *certFile)
		}
	}

	if len(nodename) > 0 && len(*nodenameFile) > 0 {
		if err := ioutil.WriteFile(*nodenameFile, []byte(nodename), os.ModePerm); err != nil {
			logger.Fatalf("Error saving nodename: %s", err.Error())
		} else {
			logger.Debugf("Nodename saved on %s", *nodenameFile)
		}
	}

	logger.Infoln("Calling finish script")
	if err := endScript(*scriptFile, *logFile); err != nil {
		logger.Error(err)
	}

	logger.Info("Halted")
	select {} // Wait forerver
}

// registrationProccess tries to register the device. I will send "register"
// requests to the server and then wait for a "registered" response containing
// an UUID. Once the UUID is obtained, if a database name is provided the
// UUID will be persisted for future requests.
func registrationProcess(apiClient *APIClient, db *Database) (uuid string, err error) {
	if db != nil {
		uuid, err = db.LoadUUID(*hash)
		logger.Info("Loading UUID from database")
		
		if err != nil {
			// logger.Error("Failed to load from DB")
			return // Error
		}
		if len(uuid) > 0 {
			logger.Debugln("Loaded UUID from database")
			return // Found UUID
		}
		logger.Debug(uuid)
	}

	for {
		logger.Debugln("Requesting new UUID")
		uuid, err = apiClient.Register()
		if err != nil {
			logger.Error("api client register")
			return
		}
		if apiClient.IsRegistered() {
			break
		}

		// Don't flood the server
		time.Sleep(time.Duration(*sleepTime) * time.Second)
	}

	if db != nil {
		db.StoreUUID(*hash, uuid)
		logger.WithField("uuid", uuid).Debugf("UUID saved to database")
	}

	return
}

// verificationProccess sends "verify" requests and waits for an "claimed"
// response. The first "claimed" response should contain a certificate and
// a node name that must be saved to disk.
func verificationProcess(uuid string, apiClient *APIClient, db *Database) (cert, nodename string, err error) {
	for {
		logger.Debugln("Requesting verification")
		err = apiClient.Verify(uuid)
		if err != nil {
			return
		}
		if apiClient.IsClaimed() {
			break
		}

		// Don't flood the server
		time.Sleep(time.Duration(*sleepTime) * time.Second)
	}

	// It is necessary to convert '\n' to actual line breaks
	// and remove the quotes
	cert = apiClient.GetCertificate()
	if len(cert) > 0 {
		cert = strings.Replace(cert, "\\n", "\n", -1)
		cert = strings.Replace(cert, `"`, ``, -1)
	}

	nodename = apiClient.GetNodename()

	return
}

func halt() {
	logger.Error("Halted")
	select {}
}
