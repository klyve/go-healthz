# go-healthz
[![Build Status](https://travis-ci.org/klyve/go-healthz.svg?branch=master)](https://travis-ci.org/klyve/go-healthz)
[![Coverage Status](https://coveralls.io/repos/github/klyve/go-healthz/badge.svg?branch=master)](https://coveralls.io/github/klyve/go-healthz?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/klyve/go-healthz)](https://goreportcard.com/report/github.com/klyve/go-healthz)

go-healthz is a `health` and `liveness` package for golang, 
it provides a quick and easy interface to add healthz services to your application.
With `Providers` you can add health checking to the critical parts of your application.
`Providers` with a detailed view will give you the current status of all health checking services.

For more information about Liveness and rediness probes visit
[https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)


## Installation 

```sh
go get -u github.com/klyve/go-healthz
```

## Usage

### healthz 
-----------------------------------------------------------------------------------------------------------
| Constructor args      | description                                                                     |
|:----------------------|:---------------------------------------------------------------------------------
| Providers             | List of Providers, these have to conform to the Checkable interface type        |
| Logger                | Logger used by healthz, if none is passed logging will be disabled              |
| Detailed              | Setting Detailed to true will enable detailed healthz service reports           |
| FailCode              | Used to set a custom HTTP status code if the healthz check fails                |
-----------------------------------------------------------------------------------------------------------

### Provider
--------------------------------------------------------------------------------------------------------------------------------------
| Provider args         | description                                                                                                |
|:----------------------|:------------------------------------------------------------------------------------------------------------
| Name                  | The name of the checkable provider (example database), this is used in errors and detailed reports         |
| Handle                | The handle is the struct handle we want to use to check for health, must conform to Checkable              |
--------------------------------------------------------------------------------------------------------------------------------------


### Use with built in server
healthz provides a simple http server for you to run the service on
If you run an existing webserver you can chose to let healthz run on it's own port and not expose it.
```go
func main() {
    logger := log.New(os.Stdout, "", 0)
    
	instance := healthz.Instance{
		Logger:    logrus.New(),
		Detailed:  true,
	}

	server := healthz.Server{
		ListenAddr:   ":3000",
		Instance:     instance,
		ServerLogger: logger,
	}

    go server.Start()

    // Other blocking code 
}
```

### Use with custom server
You can easily run this with a custom webserver,
go-healthz instance provides the functions `Healthz()` and `Liveness()`
These functions return `http.HandlerFunc`
```go
func main() {
    logger := log.New(os.Stdout, "", 0)
    
	instance := healthz.Instance{
		Logger:    logrus.New(),
		Detailed:  true,
    }
    
    // Some arbitrary server
    server.Get("/healthz", instance.Healthz())
    server.Get("/liveness", instance.Liveness())
}
```

### Creating a checkable type
A checkable type can be anything as long as it has the function `Healthz()` accessible.
If this function return nil we assume it's healthy and fine. 

```go
type healthProvider struct {}

// Required on the object type
func (h healthProvider) Healthz() error {
    return errors.New("Will always fail")
}
```

### Adding providers
A provider is an object that can report Health back to the healthz package.

```go
func main() {
    logger := log.New(os.Stdout, "", 0)
    
	instance := healthz.Instance{
        Logger:    logrus.New(),
        Providers: []healthz.Provider{
	        healthz.Provider{
		        Handle: healthProvider{},
		        Name:   "myProvider",
	        },
        }
    }
    
    // Some arbitrary server
    server.Get("/healthz", instance.Healthz())
    server.Get("/liveness", instance.Liveness())
}
```


### Response values

#### Healthz
healthz will return the health values with `errors` if they are present
The returned values differ if `Detailed` is set to true in the instance

##### Passing

Passing healthz checks return status code `200 OK`

Without detailed view
```json
{
  "healthy": true
}
```

With `Detailed` set 
```json
{
  "services": [{ "name": "myProvider", "healthy": true }],
  "healthy": true
}
```

##### Failing

Failing healthz checks return status code `503 SERVICE UNAVAILABLE`
Can be overridden by setting `FailCode` to a value in the instance definition.

Without detailed view
```json
{
  "errors": [{ "name": "myProvider", "message": "Unable to do something" }],
  "healthy": false
}
```
With `Detailed` set 
```json
{
  "services": [{ "name": "myProvider", "healthy": false }],
  "errors": [{ "name": "myProvider", "message": "Unable to do something" }],
  "healthy": false
}
```

#### Liveness

Liveness will always return `200 OK` as long as it's running
```json
OK
```