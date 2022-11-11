package metrics

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
	serverhttp "github.com/luraproject/lura/v2/transport/http/server"
	server "github.com/luraproject/lura/v2/transport/http/server/plugin"
)

func ExampleRegister() {
	cfg := config.ServiceConfig{}
	logger := logging.NoOp

	Register(context.Background(), cfg.ExtraConfig, logger)

	backendFactory := proxy.HTTPProxyFactory(&http.Client{})
	backendFactory = BackendFactory("backend", backendFactory)

	pf := proxy.NewDefaultFactory(backendFactory, logger)
	pf = ProxyFactory("proxy", pf)

	handlerFactory := router.CustomErrorEndpointHandler(logger, serverhttp.DefaultToHTTPError)
	handlerFactory = HandlerFactory(handlerFactory)

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
