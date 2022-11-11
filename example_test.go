package metrics_test

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

func ExampleRegister() {
	cfg := config.ServiceConfig{
		ExtraConfig: map[string]interface{}{
			"github_com/jbactad/krakend_newrelic_v2": map[string]interface{}{
				"rate": 100,
			},
		},
	}
	logger := logging.NoOp

	metrics.Register(context.Background(), cfg.ExtraConfig, logger)

	backendFactory := proxy.HTTPProxyFactory(&http.Client{})
	backendFactory = metrics.BackendFactory("backend", backendFactory)

	pf := proxy.NewDefaultFactory(backendFactory, logger)
	pf = metrics.ProxyFactory("proxy", pf)

	handlerFactory := router.CustomErrorEndpointHandler(logger, serverhttp.DefaultToHTTPError)
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
