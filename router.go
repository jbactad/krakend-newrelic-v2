package metrics

import (
	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

var ginMiddlewareProvider = func(application NRApplication) gin.HandlerFunc {
	nrApp, ok := application.(*newrelic.Application)
	if !ok {
		return emptyMW
	}

	return nrgin.Middleware(nrApp)
}

// Middleware adds NewRelic middleware
func Middleware() gin.HandlerFunc {
	if app == nil {
		return emptyMW
	}

	if app.Config.InstrumentationRate == 0 {
		return emptyMW
	}

	nrMiddleware := ginMiddlewareProvider(app.NRApplication)

	if app.Config.InstrumentationRate == 100 {
		return nrMiddleware
	}

	rate := float64(app.Config.InstrumentationRate) / 100.0

	return ratedMW(nrMiddleware, rate)
}

// HandlerFactory includes NewRelic transaction specific configuration endpoint naming
func HandlerFactory(handlerFactory router.HandlerFactory) router.HandlerFactory {
	if app == nil {
		return handlerFactory
	}
	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		handler := handlerFactory(cfg, p)
		return func(ctx *gin.Context) {
			txn := app.TransactionManager.TransactionFromContext(ctx)
			if txn != nil {
				txn.SetName(cfg.Endpoint)
			}

			handler(ctx)
		}
	}
}

func ratedMW(middleware gin.HandlerFunc, rate float64) gin.HandlerFunc {
	next := make(chan float64, 1000)
	go func(out chan<- float64) {
		for {
			out <- rand.Float64()
		}
	}(next)

	return func(c *gin.Context) {
		if n := <-next; n <= rate {
			middleware(c)
			return
		}
		emptyMW(c)
	}
}

func emptyMW(c *gin.Context) {
	c.Next()
}
