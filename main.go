package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/capnm/sysinfo"
	_ "github.com/mattn/go-sqlite3"
)

var version string

var (
	debug         *bool       // Debug flag
	url           *string     // API url
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
	scriptLogFile = flag.String("script-log", "/var/log/rb-register/finish.log", "Log to save the result of the script called")
	debug = flag.Bool("debug", false, "Show debug info")
	url = flag.String("url", "http://localhost", "Protocol and hostname to connect")
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

	// Check mandatory arguments
	if len(*deviceAlias) == 0 {
		flag.Usage()
		logger.Fatal("You must provide a device alias")
	}
}

func main() {
	// Handle Ctrl+C

	// Get the type of the device as an integer value
	deviceType, err := getDeviceType(*deviceAlias)
	if err != nil {
		logger.Fatal("Invalid device alias")
	}

	// Obtain information of the system
	si = sysinfo.Get()

	// Daemonize the application
	if *daemonFlag {
		daemonize()
	}

	// Initialize database
	db := NewDatabase(DatabaseConfig{dbFile: *dbFile})
	defer db.Close()

	// Create a new API client for handle the connection with the API
	apiClient := NewAPIClient(
		APIClientConfig{
			URL:        *url,
			Hash:       *hash,
			Cpus:       runtime.NumCPU(),
			Memory:     si.TotalRam,
			DeviceType: deviceType,
		},
	)

	logger.Info("Start the registration process")
	uuid := registrationProcess(apiClient, db)

	logger.Infof("Using UUID: %s", uuid)

	logger.Info("Start the verification process")
	cert, nodename := verificationProcess(uuid, apiClient, db)

	// Save the certificate to a file
	if len(cert) > 0 && certFile != nil {
		if err := ioutil.WriteFile(*certFile, []byte(cert), os.ModePerm); err != nil {
			logger.Fatalf("Error saving certificate: %s", err.Error())
		}
	}

	// Get nodename and save it to a file
	if len(*nodenameFile) > 0 {
		if err := ioutil.WriteFile(*nodenameFile, []byte(nodename), os.ModePerm); err != nil {
			logger.Fatalf("Error saving nodename: %s", err.Error())
		}
	}

	// Call the finish script
	if err := endScript(*scriptFile, *logFile); err != nil {
		logger.Error(err)
	}
	logger.Infof("Chef called")

	// Wait for SIGINT
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	<-ctrlc
}

// registrationProccess tries to register the device. I will send "register"
// requests to the server and then wait for a "registered" response containing
// an UUID. Once the UUID is obtained, if a database name is provided the
// UUID will be persisted for future requests.
func registrationProcess(apiClient *APIClient, db *Database) string {
	var uuid string
	var err error

	// First check if already exists an UUID for the given HASH stored on DB so
	// it is not necessary to send any "register" request
	if db != nil {
		uuid, err = db.LoadUUID(*hash)

		if err != nil || len(uuid) <= 0 {
			logger.Warn("Can't load UUID from database")
		} else {
			logger.Debugf("Load UUID from DB: %s", uuid)
			return uuid
		}
	}

	// No UUID could be load from DB, send register messages until a "registered"
	// response arrives
	for {
		if err = apiClient.Register(); err != nil {
			logger.Warnf("Error registering device: %s", err)
			break
		}
		if apiClient.IsRegistered() {
			break
		}

		// Don't flood the server
		time.Sleep(time.Duration(*sleepTime) * time.Second)
	}

	// Once an UUID has been obtained, if a database has been provided then
	// persist the UUID
	if db != nil {
		uuid, err = apiClient.GetUUID()
		if err != nil {
			logger.Fatalf("Error getting UUID: %s", err)
		}

		db.StoreUUID(*hash, uuid)
	}

	return uuid
}

// verificationProccess sends "verify" requests and waits for an "claimed"
// response. The first "claimed" response should contain a certificate and
// a node name that must be saved to disk.
func verificationProcess(uuid string, apiClient *APIClient, db *Database) (cert, nodename string) {
	var err error

	for {
		if err = apiClient.Verify(); err != nil {
			logger.Warnf("Error verifying device: %s", err)
			break

		}
		if apiClient.IsClaimed() {
			break
		}

		// Don't flood the server
		time.Sleep(time.Duration(*sleepTime) * time.Second)
	}

	// Get the certificate. It is necessary to convert '\n' to actual line breaks
	// and remove the quotes
	cert, err = apiClient.GetCertificate()
	if err != nil {
		logger.Fatalf("Error getting certificate: %s", err)
	}
	cert = strings.Replace(cert, "\\n", "\n", -1)
	cert = strings.Replace(cert, `"`, ``, -1)

	nodename, err = apiClient.GetNodename()
	if err != nil {
		logger.Fatalf("Error getting nodename: %s", err)
	}

	return
}
