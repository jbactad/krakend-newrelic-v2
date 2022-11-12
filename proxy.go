package metrics

import (
	"context"
	"fmt"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
)

// ProxyFactory creates an instrumented proxy factory
func ProxyFactory(segmentName string, next proxy.Factory) proxy.FactoryFunc {
	if app == nil {
		return next.New
	}
	return proxy.FactoryFunc(
		func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
			next, err := next.New(cfg)
			if err != nil {
				return proxy.NoopProxy, err
			}
			return NewProxyMiddleware(fmt.Sprintf("(%s) %s", segmentName, cfg.Endpoint))(next), nil
		},
	)
}

// NewProxyMiddleware adds NewRelic segmentation
func NewProxyMiddleware(segmentName string) proxy.Middleware {
	if app == nil {
		return proxy.EmptyMiddleware
	}
	return func(next ...proxy.Proxy) proxy.Proxy {
		if len(next) > 1 {
			panic(proxy.ErrTooManyProxies)
		}
		if len(next) == 0 {
			panic(proxy.ErrNotEnoughProxies)
		}
		return func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
			tx := app.TransactionManager.TransactionFromContext(ctx)
			if tx == nil {
				return next[0](ctx, req)
			}

			segment := tx.StartSegment(segmentName)
			resp, err := next[0](ctx, req)
			defer segment.End()

			return resp, err
		}
	}
}
