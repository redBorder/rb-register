package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/capnm/sysinfo"

	"redborder/rb-register-2/client"
)

var (
	debug      *bool
	url        *string // API url
	hash       *string // Required hash to perform the registration
	sleepTime  *int    // Time between requests
	deviceType *int    // Type of the requesting device
	insecure   *bool   // If true, skip SSL verification
	certPath   *string // Path to store de certificate

	si  sysinfo.SI
	log *logrus.Logger
)

// init parses flags
func init() {
	debug = flag.Bool("-debug", false, "Show debug info")
	url = flag.String("-url", "http://localhost", "Protocol and hostname to connect")
	hash = flag.String("-hash", "00000000-0000-0000-0000-000000000000", "Hash to use in the request")
	sleepTime = flag.Int("-sleep", 300, "Time between requests in seconds")
	deviceType = flag.Int("-device", 300, "Time between requests in seconds")
	insecure = flag.Bool("-no-check-certificate", false, "Dont check if the certificate is valid")
	certPath = flag.String("-cert", "/opt/rb/etc/chef/client.pem", "Certificate path")

	flag.Parse()
}

func main() {

	// Create new logger
	log = logrus.New()
	if *debug {
		log.Level = logrus.DebugLevel
	}

	// Create a new API client config for handle the connection with the API
	apiClient := client.NewApiClient(
		client.Config{
			Url:        *url,
			Hash:       *hash,
			SleepTime:  *sleepTime,
			Cpus:       runtime.NumCPU(),
			Memory:     si.TotalRam,
			DeviceType: *deviceType,
			Insecure:   *insecure,
			Debug:      *debug,
			HttpClient: &http.Client{},
		})

	// Try to register with the API
	log.Infof("Registering")
	if err := apiClient.Register(); err != nil {
		log.Fatalf("Error registering device: %s", err)
	}
	log.Infof("Registered!")

	// Start verification process. Finish when the device is claimed
	log.Infof("Verifying")
	if err := apiClient.Verify(); err != nil {
		log.Fatalf("Error verifiyin device: %s", err)
	}
	log.Infof("Verified!")

	cert, err := apiClient.GetCertificate()
	if err != nil {
		log.Errorf("Error getting certificate: %s", err)
	}

	if err := ioutil.WriteFile(*certPath, []byte(cert), os.ModePerm); err != nil {
		log.Errorf("Error saving certificate: %s", err.Error())
	}

	startChef()
}

func startChef() {
	var out bytes.Buffer

	cmd := `echo 'sh /opt/rb/bin/rb_register_finish.sh >> /var/log/rb-register/finish.log' | at now`
	exe := exec.Command("sh", "-c", cmd)
	exe.Stdout = &out

	err := exe.Run()
	if err != nil {
		log.Fatal(err)
	}
}
