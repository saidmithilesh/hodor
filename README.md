# hodor

**Hold The Door!** Hodor is a high performance, scalable API Gateway designed to keep the bad forces from reaching your services.

![Hold the door](hold_the_door.gif)

## Core philosophy

User experience of setting up and using a tool matters! Let engineers and system admins focus more on the solution and less on tedious configuration. Hodor is designed with the following tenets in mind:

- Make it super simple to configure
- Allow global as well as local customisation
- Make it easy to deploy, monitor and scale the solution
- Make it easy to log and trace issues

## Architecture

![Architecture](hodor_architecture.jpeg)

## Configuration

Let's face it. No matter how much we hate configuration, it is an important part of setting up a tool. Hodor tries to simplify the process of configuration as much as possible so that you can get done with it and move on.

Think of Hodor's configuration as comprising of two simple levels:

- Gateway wide configuration
- List of endpoint integrations where each has its own configuration

```yaml
## YAML Config
gateway:
  name: "test_gateway"
  description: "Testing API Gateway integrations"
  
  # Specify the port on which you want the gateway application to run
  port: ":443"
  
  # Set to true if you need HTTPS support
  enable_TLS: false

  # Paths to your certificate and key files
  TLS_cert: "/path/to/samplecertificate.crt"
  TLS_key: "/path/to/samplecertificate.key"
  
  # Logging
  log_level: "INFO"
  log_output: "STDIO" # also supports ELK, FILE, KAFKA
  log_credentials: "log_credentials.yml"
  
  # Total number of requests allowed per unit time and the penalty to levy when rate exceeds the specified limit
  rate_limit:
    requests: 10
    window: "10S"
    penalty: "10S"

  # CORS setup
  cors:
    # Methods you want to allow access to
    allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS

     # List of all domains you wish to allow requests from
    allowed_domains:
    - localhost:1337

    # List of headers you wish to allow (case insensitive)
    allowed_headers:
    - Content-Type
    - Authorization
    - x-your-custom-header

    # List of headers you wish to expose in the response
    exposed_headers:
    - x-custom-response-header

  endpoints:
  - name: "Get Orders"
    description: "Get a list of all orders placed in the last 24 hours by a particular customer"

    method: "GET"
    path: "/customer/:customerId/orders"

    # Optional list of middleware endpoints which you can use to add
    # custom logic on top of each endpoint
    middleware:
    - "http://yourmiddleware.com/middleware1"
    - "http://yourmiddleware.com/middleware2"

    # Specify the application backend that will handle and process the request
    backend: "http://yourbackend-api.com"

  - name: "Create Order"
    description: "Create a new order for a customer"
    method: "POST"
    path: "/customer/:customerId/order"
    middleware:
    - "http://yourmiddleware.com/middleware1"
    backend: "http://some-other-backend.com"
```

## Features

### 1. TLS Support

Transport layer security is supported out of the box. Should you need to protect your API Gateway with TLS certificates, simply set the ```enable_TLS``` flag to ```true``` and provide the paths to your certificate and key files.

```yaml
gateway:
  enable_TLS: true
  TLS_cert: "/path/to/samplecertificate.crt"
  TLS_key: "/path/to/samplecertificate.key"
```

### 2. Authentication and authorization

Hodor supports simple token based, token + secret based and JWT based auth. If you need to implement a different auth mechanism not covered under these, check the middleware section below to understand how you can integrate your own custom auth.

### 3. Customisable endpoints

You can specify as many API endpoints under an API Gateway as you need and specify which backend it should proxy the request to.
If you need to proxy requests on different endpoints to different backends, Hodor allows you to do so. You don't need to stick to just one backend for all endpoints.
Endpoint support dynamic paths with variables in them so that you don't have to give up your clean routes.

```yaml
gateway:
  # ...
  endpoints:
  - name: "Get Orders"
    description: "Get a list of all orders placed in the last 24 hours by a particular customer"

    method: "GET"

    # Dynamic path where customerId is a variable
    path: "/customer/:customerId/orders"

    # ...

    backend: "http://yourbackend-api.com"
```

### 4. Rate Limiting

API call rate limits can be applied on an overall gateway level (requests across all endpoints combined)

```yaml
gateway:
  # ...
  # Allow a maximum of 10 requests in 1 second
  # but levy no penalty if rate exceeds the limit
  rate_limit:
    requests: 10
    window: "1S" # 1S = 1 second
    penalty: "-" # - indicates no penalty
  # ...
```

as well as on each endpoint individually

```yaml
gateway:
  # ...
  endpoints:
  - name: "Get orders"
    # ...
    # Allow a maximum of 30 requests in 10 minutes
    # and levy a penalty of 20 minutes if rate exceeds the limit
    rate_limit:
      requests: 30
      window: "10M"
      penalty: "20M"
    # ...
```

- If rate limits are present on both gateway level as well as endpoint level, endpoint rate limit takes precedence
- If no rate limit is specified at endpoint level, it defaults to the gateway level rate limit which applies on requests across all endpoints combined
- Hodor uses a sliding window protocol on top of redis to implement rate limiting
- If number of requests made in a window exceed the limit or if a penalty is levied, the gateway returns ```429 Too many requests``` HTTP status to the client
- `window` and `penalty` fields accept durations in seconds `1S = 1 second`, minutes `5M = 5 minutes` and hours `0.5H = half an hour` format

### 5. CORS

- If your API is going to be accessed from web clients, you might need cross origin support
- Similar to rate-limiting, CORS settings can also be enabled on gateway level as well as individual endpoint level
- Endpoint config takes precedence when both are present
- Sample CORS config

```yaml
gateway:
  # ...
  cors:
    # White list the methods you want to allow access to
    allowed_methods:
    - "GET"
    - "POST"
    - "OPTIONS"

     # White list of all domains you wish to allow requests from
    allowed_domains:
    - "localhost:1337"
    - "yourwebsite.com"

    # List of headers you wish to allow (case insensitive)
    allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "x-your-custom-header"

    # List of headers you wish to expose in the response
    exposed_headers:
    - "x-custom-response-header"
```

### 6. Middleware

- Don't like some of the functionality Hodor provides out of the box? Or maybe you want to add some custom logic of your own. For each endpoint you can specify a list of custom HTTP/S middleware
- Whenever a request is received on an endpoint, Hodor will first proxy it in series to each middleware before proxying it to the actual backend
- If it receives a non `2xx` status code from any of the middleware, it will simply return that status code and response to the client and will not forward the request to the backend
- Only on receiving a `2xx` from all the middleware will a request be deemed eligible for proxying to the backend
- Ex: If you are trying to implement a custom auth strategy, let's say you deploy it on `http://your-custom-middleware.com/middleware1` and you actual backend application which will serve the request is on `http://your-backend.com/apiendpoint`. Hodor will first proxy the request to the middleware endpoint. If the middleware returns a `401 Unauthorized` response, it will be sent as it is to the client and the request will not be forwarded to the actual backend service.

### 7. Logging

- Hodor supports standard log levels of `DEBUG`, `INFO`, `WARNING` and `ERROR` in increasing level of priority
- Log streams can be written to file, Standard IO, ELK stack or Kafka

### 8. Easy deployment

- Hodor is built in golang. You can build a binary for any target operating system (Mac OS, Linux, Windows) and run it with a simple command `./hodor -config=/path/to/config.yml`

### 9. High performance

- Golang allows Hodor to support high number of concurrent requests with minimal memory overhead and scales well on multi-core CPUs
- #TODO: benchmark results

### (Upcoming)

- Support multiple protocols (HTTP/S, HTTP2, WebSockets)
- Usage monitoring and auto scaling support for multiple cloud providers (a.k.a Bran, the one who can warg and control Hodor)
- Load balancing between multiple backends
