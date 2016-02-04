package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/capnm/sysinfo"
	"github.com/codeskyblue/go-sh"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sevlyar/go-daemon"

	"redborder/rb-register-2/client"
	"redborder/rb-register-2/database"
)

const (
	NODE_NAME = "/opt/rb/etc/chef/nodename"
)

var (
	debug      *bool
	url        *string // API url
	hash       *string // Required hash to perform the registration
	sleepTime  *int    // Time between requests
	deviceType int     // Type of the requesting device
	insecure   *bool   // If true, skip SSL verification
	certFile   *string // Path to store de certificate
	dbFile     *string // File to persist the state
	daemonFlag *bool   // Start in daemon mode
	pid        *string // Path to PID file
	logFile    *string // Log file

	si  *sysinfo.SI
	log *logrus.Logger
)

var deviceAlias = map[string]int{
	"AP":          20,
	"PROXY":       31,
	"IPS":         32,
	"IPS_GENERIC": 33,
}

// init parses flags
func init() {
	debug = flag.Bool("debug", false, "Show debug info")
	url = flag.String("url", "http://localhost", "Protocol and hostname to connect")
	hash = flag.String("hash", "00000000-0000-0000-0000-000000000000", "Hash to use in the request")
	sleepTime = flag.Int("sleep", 300, "Time between requests in seconds")
	deviceTypeFlag := flag.String("type", "", "Type of the registering device")
	insecure = flag.Bool("no-check-certificate", false, "Dont check if the certificate is valid")
	certFile = flag.String("cert", "/opt/rb/etc/chef/client.pem", "Certificate file")
	dbFile = flag.String("db", "", "File to persist the state")
	daemonFlag = flag.Bool("daemon", false, "Start in daemon mode")
	pid = flag.String("pid", "pid", "File containing PID")
	logFile = flag.String("log", "log", "Log file")

	flag.Parse()

	// Create new logger
	log = logrus.New()
	if *debug {
		log.Level = logrus.DebugLevel
	}

	// Check device type arg
	if len(*deviceTypeFlag) == 0 {
		flag.Usage()
		log.Fatal("You must provide a device type")
	}

	var err error
	if deviceAlias[*deviceTypeFlag] != 0 {
		deviceType = deviceAlias[*deviceTypeFlag]
	} else {
		deviceType, err = strconv.Atoi(*deviceTypeFlag)
		if err != nil {
			log.Fatal("Invalid device type")
		}
	}

	si = sysinfo.Get()
}

func main() {

	// Daemonize the application
	if *daemonFlag {
		cntxt := &daemon.Context{
			PidFileName: *pid,
			PidFilePerm: 0644,
			LogFileName: *logFile,
			LogFilePerm: 0640,
			WorkDir:     "./",
			Umask:       027,
			Args:        os.Args,
		}

		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatalln(err)
		}
		if d != nil {
			log.Infof("Daemon started [PID: %d]", d.Pid)
			return
		}
		defer cntxt.Release()
	}

	var db *database.Database

	// Load database if neccesary
	if dbFile != nil {
		sqldb, err := sql.Open("sqlite3", *dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer sqldb.Close()

		db = database.NewDatabase(sqldb)
	}

	// Create a new HTTP Client
	var httpClient *http.Client
	if *insecure {
		httpClient = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	} else {
		httpClient = &http.Client{}
	}

	// Create a new API client config for handle the connection with the API
	apiClient := client.NewApiClient(
		client.Config{
			Url:        *url,
			Hash:       *hash,
			Cpus:       runtime.NumCPU(),
			Memory:     si.TotalRam,
			DeviceType: deviceType,
			Debug:      *debug,
		}, httpClient, db)

	// Check if exists an UUID stored on DB
	if db != nil && apiClient.IsRegistered() {
		log.Infof("Loaded UUID from DB")
	} else {
		// No previous UUID, try to register
		for {
			if err := apiClient.Register(); err != nil {
				log.Fatalf("Error registering device: %s", err)
			}

			if apiClient.IsRegistered() {
				log.Infof("Registered!")
				break
			} else {
				time.Sleep(time.Duration(*sleepTime) * time.Second)
			}
		}
	}

	// Start verification process. Finish when the device is claimed
	log.Infof("Requesting certificate")
	for !apiClient.IsClaimed() {
		if err := apiClient.Verify(); err != nil {
			log.Fatalf("Error registering device: %s", err)
		}

		if !apiClient.IsClaimed() {
			time.Sleep(time.Duration(*sleepTime) * time.Second)
		} else {
			log.Infof("Claimed!")
		}
	}

	// Get the certificate
	cert, err := apiClient.GetCertificate()
	if err != nil {
		log.Errorf("Error getting certificate: %s", err)
	}
	cert = strings.Replace(cert, "\\n", "\n", -1)
	cert = strings.Replace(cert, `"`, ``, -1)

	// Save the certificate to a file
	if err := ioutil.WriteFile(*certFile, []byte(cert), os.ModePerm); err != nil {
		log.Errorf("Error saving certificate: %s", err.Error())
	}

	// Get nodename and save it to a file
	if nodename, err := apiClient.GetNodename(); err != nil {
		if err := ioutil.WriteFile(NODE_NAME, []byte(nodename), os.ModePerm); err != nil {
			log.Errorf("Error saving nodename: %s", err.Error())
		}
	}

	// Launch chef client
	sh.Echo("bash /opt/rb/bin/rb_register_finish.sh >> /var/log/rb-register/finish.log").Command("at", "now").Run()
	log.Infof("Chef called")

	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)

	<-ctrlc
}
