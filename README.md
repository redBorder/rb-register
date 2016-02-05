# rb_register

## Params

Usage of **rb_register** and default values:

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
-sleep int
      Time between requests in seconds (default 300)
-type string
      Type of the registering device
-url string
      Protocol and hostname to connect (default "http://localhost")
```

## Description

The status of the sensor can be:

- **Not registered**: The sensor doesn't appear at the cloud and it has never contacted with it. This is the initial status for this sensor.
- **Registered**: The sensor is registered at the cloud but nobody claimed it.
- **Claimed**: The sensor registered and claimed by the client but need final step (download certificate)

### Unregistered status

The sensor will start at "Not registered" status so it will try to contact to given url (managed by the same process than `cloud.redborder.net` now). The sensor will generate an http post message sending this data:

#### Register request

```javascript
{
    "order": "register",
    "cpu": /* Number of CPUs */,
    "memory": /* Memory avalable */,
    "type": /* Type of sensor */,
    "hash": /* HASH */
}
```

#### Register response

If not registered, the cloud generates a **new** `UUID` and returns it in a message.

  ```javascript
  {
    "status": "registered",
    "mac": /* HASH */,
    "uuid": /* UUID */
  }
  ```


### Registered status

Afther the sensor receives the "registered" status, it will send "verify" request instead of "register" request. A verify request expects a certificate.

#### Verify request

```javascript
{
    "order": "verify",
    "mac": /* MAC address */,
    "uuid": /* UUID */
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
    "status": "claimed",
    "cert": /* CERTIFICARTE */,
    "nodename": /* The name of the node */
}
```

### Claimed status

When a "claimed" status is received, the certificate is saved on the machine and the the app halts util is stopped.