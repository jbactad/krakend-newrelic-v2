package metrics

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
)

// BackendFactory creates an instrumented backend factory
func BackendFactory(segmentName string, next proxy.BackendFactory) proxy.BackendFactory {
	if app == nil {
		return next
	}

	return func(cfg *config.Backend) proxy.Proxy {
		return NewBackend(segmentName, next(cfg))
	}
}

// NewBackend includes NewRelic segmentation
func NewBackend(segmentName string, next proxy.Proxy) proxy.Proxy {
	if app == nil {
		return next
	}

	return func(ctx context.Context, proxyReq *proxy.Request) (*proxy.Response, error) {
		tx := app.TransactionManager.TransactionFromContext(ctx)
		if tx == nil {
			return next(ctx, proxyReq)
		}

		req, err := toHttpRequest(proxyReq)
		if err != nil {
			return nil, err
		}

		externalSegment := app.TransactionManager.StartExternalSegment(tx, req)
		proxyReq.Headers = req.Header
		defer func() {
			externalSegment.End()
			if req.Body != nil {
				req.Body.Close()
			}
		}()

		resp, err := next(ctx, proxyReq)
		if err == nil {
			externalSegment.SetStatusCode(resp.Metadata.StatusCode)
		}

		return resp, err
	}
}

func toHttpRequest(req *proxy.Request) (*http.Request, error) {
	requestToBackend, err := http.NewRequest(strings.ToTitle(req.Method), req.URL.String(), req.Body)
	if err != nil {
		return nil, err
	}
	requestToBackend.Header = make(map[string][]string, len(req.Headers))
	for k, vs := range req.Headers {
		tmp := make([]string, len(vs))
		copy(tmp, vs)
		requestToBackend.Header[k] = tmp
	}
	if req.Body != nil {
		if v, ok := req.Headers["Content-Length"]; ok && len(v) == 1 && v[0] != "chunked" {
			if size, err := strconv.Atoi(v[0]); err == nil {
				requestToBackend.ContentLength = int64(size)
			}
		}
	}

	return requestToBackend, nil
}
