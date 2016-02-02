# rb_register

## Params

```
  --help                 Display this usage information.
  
  ACTIONS:
  --register             Perform registration process.
  
  OPTIONS:
  --url                  Protocol and hostname to connect.
  --iface                Network if to get the MAC address.
  --sleep                Time between requests in seconds.
  --type                 ips | proxy | ap
  --pid                  Path to PID file
  --no-check-certificate Dont check if the certificate is valid
  --daemonize            Start as daemon
```

- `DEFAULT HOST`: `https://register.app.redborder.net`
- `DEFAULT IFACE`: `eth0`
- `DEFAULT PID FILE`: `/tmp/rb_register.pid`
- `DEFAULT CERTIFICATE PATH`: `/opt/rb/etc/chef/client.pem`
- `DEFAULT_SLEEP_TIME`: `5`

## Description

There are three types of main sensors now:
- Proxy
- IPS
- AP

Sensors should be claimed by the client

The status of the sensor can be:

- **Not registered**: The sensor doesn't appear at the cloud and it has never contacted with it. This is the initial status for this sensor.
- **Registered**: The sensor is registered at the cloud but nobody claimed it.
- **Claimed**: The sensor registered and claimed by the client but need final step (download certificate)

### Unregistered status

The sensor will start at "Not registered" status so it will try to contact to `register.redborder.net` (managed by the same process than `cloud.redborder.net` now). The sensor will generate an http post message sending this data:

#### Register request

```javascript
{
    "order": "register",
    "cpu": /* Number of CPUs */,
    "memory": /* Memory avalable */,
    "type": /* Type of sensor (2(IPS) | 10(PROXY) | AP) */,
    "mac": /* MAC address */
}
```

#### Register response

If not registered, the cloud generates a **new** `UUID` and returns it in a message.

  ```javascript
  {
    "status": "registered",
    "mac": /* MAC address */,
    "uuid": /* UUID */
  }
  ```

_NOTE: The sensor can be previously registered but when the application starts, it always starts at "unregistered" status so it will send the register request. The server should know about this and send the **previous** `UUID` on the response instead of generate a new one everytime._

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
    "cert": /* CERTIFICARTE */
}
```

### Claimed status

When a "claimed" status is received, the certificate is saved on the machine and the the app halts.