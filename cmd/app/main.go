package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"redborder/rb-register-2/client"
)

const (
	AP          = 20
	PROXY       = 31
	IPS         = 32
	IPS_GENERIC = 33
)

var (
	log        *logrus.Logger
	debug      bool
	configFile string
)

type Config struct {
	Url        string `yaml:"url"`      // API url
	Hash       string `yaml:"hash"`     // Required hash to perform the registration
	SleepTime  int    `yaml:"sleep"`    // Time between requests
	Cpus       int    `yaml:"cpus"`     // Number of CPU of the computer
	Memory     int64  `yaml:"memory"`   // Amount of memory of the computer
	DeviceType int    `yaml:"type"`     // Type of the requesting device
	Insecure   bool   `yaml:"insecure"` // If true, skip SSL verification
	CertPath   string `yaml:"certpath"` // If true, skip SSL verification
}

// init parses flags and initializes logger
func init() {
	configFileFlag := flag.String("config", "", "Config file")
	debugFlag := flag.Bool("debug", false, "Show debug info")

	flag.Parse()

	debug = *debugFlag

	log = logrus.New()
	if debug {
		log.Level = logrus.DebugLevel
	}

	if len(*configFileFlag) == 0 {
		flag.Usage()
		log.Fatal("No config file provided")
	}
	configFile = *configFileFlag
}

func main() {

	// Load the config file
	config, err := loadConfigFile(configFile)
	if err != nil {
		log.Fatalf("Error reading configuration file: %s", err.Error())
	}

	// Create a new API client config for handle the connection with the API
	apiClient := client.NewApiClient(
		client.Config{
			Url:        config.Url,
			Hash:       config.Hash,
			SleepTime:  config.SleepTime,
			Cpus:       config.Cpus,
			Memory:     config.Memory,
			DeviceType: config.DeviceType,
			Insecure:   config.Insecure,
			Debug:      debug,
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

	if err := ioutil.WriteFile(config.CertPath, []byte(cert), os.ModePerm); err != nil {
		log.Errorf("Error saving certificate: %s", err.Error())
	}
}

// loadConfigFile opens a configuration YAML file and parse the content to a
// structure.
func loadConfigFile(configFile string) (config Config, err error) {
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	err = yaml.Unmarshal([]byte(configData), &config)
	if err != nil {
		return
	}

	return
}
