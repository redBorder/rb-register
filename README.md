[![Build Status](https://travis-ci.org/redBorder/rb-register.svg?branch=master)](https://travis-ci.org/redBorder/rb-register)
[![Coverage Status](https://coveralls.io/repos/github/redBorder/rb-register/badge.svg?branch=master)](https://coveralls.io/github/redBorder/rb-register?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/redBorder/rb-register)](https://goreportcard.com/report/github.com/redBorder/rb-register)

# rb-register

Application written in GO that allow sensors to be registered by the redBorder
Live Cloud.

## Installing

To install this application ensure you have the `GOPATH` environment variable
set and **[glide](https://glide.sh/)** installed.

```bash
curl https://glide.sh/get | sh
```

And then:

1. Clone this repo and cd to the project

    ```bash
    git clone https://github.com/redBorder/rb-register.git && cd rb-register
    ```
2. Install dependencies and compile

    ```bash
    make
    ```
3. Install on desired directory

    ```bash
    prefix=/opt/rb make install
    ```

## Usage

Usage of **rb-register** and default values:

```
-cert string
  	Certificate file (default "/opt/rb/etc/chef/client.pem")
-daemon
  	Start in daemon mode
-db string
  	File to persist the state
-debug
  	Show debug info
-hash string
  	Hash to use in the request (default "00000000-0000-0000-0000-000000000000")
-log string
  	Log file (default "log")
-no-check-certificate
  	Dont check if the certificate is valid
-nodename string
  	File to store nodename
-pid string
  	File containing PID (default "pid")
-script string
  	Script to call after the certificate has been obtained (default "/opt/rb/bin/rb_register_finish.sh")
-script-log string
  	Log to save the result of the script called (default "/var/log/rb-register/finish.log")
-sleep int
  	Time between requests in seconds (default 300)
-type string
  	Type of the registering device
-url string
  	Protocol and hostname to connect (default "http://localhost")
-version
  	Display version
```

## Description

The status of the sensor can be:

- **Not registered**: The sensor doesn't appear at the cloud and it has never
contacted with it. This is the initial status for this sensor.
- **Registered**: The sensor is registered at the cloud but nobody claimed it.
- **Claimed**: The sensor registered and claimed by the client but need final
step (download certificate).

### Register process

The sensor will start at "Not registered" status so it will try to contact
to given url (managed by the same process than `live.redborder.com` now). The
sensor will generate an http post message sending this data:

#### Register request

```javascript
{
    "order":  "register",
    "cpu":    /* Number of CPUs   */,
    "memory": /* Memory avalable  */,
    "type":   /* Type of sensor   */,
    "hash":   /* HASH             */
}
```

#### Register response

If not registered, the cloud generates a **new** `UUID` and returns it in a
message. If a database is provided then the application will persist the UUID
so its not necessary to send the `register` request everytime the application
starts.

  ```javascript
  {
    "status": "registered",
    "mac":  /* HASH */,
    "uuid": /* UUID */
  }
  ```

### Verification process

After the sensor receives the "registered" status, it will send `verify`
requests instead of `register` request. A verify request expects a `claimed`
response along with a certificate and node name.

#### Verify request

```javascript
{
    "order": "verify",
    "mac":   /* MAC address */,
    "uuid":  /* UUID        */
}
```

#### Verify response

When the sensor sends a "verify" request expects a certificate, but if the sensor hasn't been claimed the certificate doesn't exists yet.

1. If the sensor hasn't been claimed and there isn't a certificate, the response should be:

```javascript
{
    "status": "registered"
}
```

2. If the sensor has been claimed and there is a certificate, the response should be:

```javascript
{
    "status":   "claimed",
    "cert":     /* CERTIFICARTE         */,
    "nodename": /* The name of the node */
}
```

### Done status

When a "claimed" status is received, the certificate and the node name are saved
on disk and the application will execute a command before it halts.
