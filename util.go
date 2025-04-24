package main

import (
	"errors"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"strconv"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	daemon "github.com/sevlyar/go-daemon"
)

// getDeviceType try to get the representation of a device alias (string) as an
// integer. First, it will try to find if the device alias is in a map, if not,
// it will try to decode as an integer (using atoi). If both methods fails it
// will return an error.
func getDeviceType(alias string) (deviceType int, err error) {
  var aliasMap = map[string]int{
      "ap":             20,
      "proxy":          31,
      "ips":            32,
      "ips-generic":    33,
      "exporter":       41,
      "intrusion-proxy": 98, 
  }

	// Check device type arg
	if len(alias) == 0 {
		logrus.Fatal("You must provide a device type")
	}

	if v := aliasMap[alias]; v != 0 {
		deviceType = v
	} else {
		deviceType, err = strconv.Atoi(alias)
		if err != nil {
			err = errors.New("Invalid device type")
		}
	}

	return
}

// daemonize let the app running on the background detached from the terminal
// and will log to syslog
func daemonize() {
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "rb-register")
	if err != nil {
		logrus.Error("Unable to connect to local syslog daemon")
	} else {
		logger.Hooks.Add(hook)
	}

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
		logger.Fatalln(err)
	}
	if d != nil {
		logger.Infof("Daemon started [PID: %d]", d.Pid)
		return
	}

	defer cntxt.Release()
}

func endScript(script, logFileName string) error {
	cmd := exec.Command(script)

	// open the out file for writing
	logfile, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	defer logfile.Close()

	cmd.Stdout = logfile
	err = cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func displayVersion() {
	fmt.Println("RB_REGISTER VERSION:\t", version)
	fmt.Println("GO VERSION:\t\t", goVersion)
}
