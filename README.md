# krakend-newrelic-v2

[![Go Reference](https://pkg.go.dev/badge/github.com/jbactad/krakend-newrelic-v2.svg)](https://pkg.go.dev/github.com/jbactad/krakend-newrelic-v2)
![Go](https://github.com/jbactad/krakend-newrelic-v2/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/jbactad/krakend-newrelic-v2/branch/main/graph/badge.svg?token=9CCWX167AA)](https://codecov.io/gh/jbactad/krakend-newrelic-v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbactad/krakend-newrelic-v2)](https://goreportcard.com/report/github.com/jbactad/krakend-newrelic-v2)

A NewRelic middleware that uses the NewRelic Go Agent v3

## Installation

```bash
go get -u github.com/jbactad/krakend-newrelic-v2
```

## Quick Start

To enable NewRelic instrumentation in your krakend api gateway,
you need to make sure and call `metrics.Register` function to initialize the `newrelic.Application` from your gateway.

There are 3 middlewares where you can enable instrumentation in your krakend api gateway, Handler, Proxy and Backend.

```go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	metrics "github.com/jbactad/krakend-newrelic-v2"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
	serverhttp "github.com/luraproject/lura/v2/transport/http/server"
	server "github.com/luraproject/lura/v2/transport/http/server/plugin"
)

func main() {
	cfg := config.ServiceConfig{}
	logger := logging.NoOp

	// Initializes the metrics collector.
	metrics.Register(context.Background(), cfg.ExtraConfig, logger)

	backendFactory := proxy.HTTPProxyFactory(&http.Client{})

	// Wrap the default BackendFactory with instrumented BackendFactory 
	backendFactory = metrics.BackendFactory("backend", backendFactory)

	pf := proxy.NewDefaultFactory(backendFactory, logger)

	// Wrap the default ProxyFactory with instrumented ProxyFactory
	pf = metrics.ProxyFactory("proxy", pf)

	handlerFactory := router.CustomErrorEndpointHandler(logger, serverhttp.DefaultToHTTPError)

	// Wrap the default HandlerFactory with instrumented HandlerFactory
	handlerFactory = metrics.HandlerFactory(handlerFactory)

	engine := gin.New()

	// setup the krakend router
	routerFactory := router.NewFactory(
		router.Config{
			Engine:         engine,
			ProxyFactory:   pf,
			Logger:         logger,
			RunServer:      router.RunServerFunc(server.New(logger, serverhttp.RunServer)),
			HandlerFactory: handlerFactory,
		},
	)

	// start the engines
	logger.Info("Starting the KrakenD instance")
	routerFactory.NewWithContext(context.Background()).Run(cfg)
}
```

Then in your gateway's config file make sure to add `github_com/jbactad/krakend_newrelic_v2` in the
service `extra_config` section.

```json
{
  "version": 3.0,
  "extra_config": {
    "github_com/jbactad/krakend_newrelic_v2": {
      "rate": 100
    }
  }
}
```

## Configuring

NewRelic related configurations are all read from environment variables.
Refer to [newrelic go agent](https://pkg.go.dev/github.com/newrelic/go-agent/v3/newrelic@v3.18.1#ConfigFromEnvironment)
package to know more of which NewRelic options can be configure.

From krakend configuration file, these are the following options you can configure.

| Name | Type | Description                                           |
|------|------|-------------------------------------------------------|
| rate | int  | The rate the middlewares instrument your application. |


## Development

### Requirements

To start development, make sure you have the following dependencies installed in your development environment.

- golang >=v1.17

### Setup

Run the following to install the necessary tools to run tests.

```bash
make setup
```

### Generate mocks for unit tests

```bash
make gen
```

### Running unit tests

To run the unit tests, execute the following command.

```bash
make test-unit
```
