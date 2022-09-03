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

	return func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
		tx := app.TransactionManager.TransactionFromContext(ctx)
		if tx == nil {
			return next(ctx, req)
		}

		requestToBackend, closer, err := toHttpRequest(req)
		if err != nil {
			return nil, err
		}

		externalSegment := app.TransactionManager.StartExternalSegment(tx, requestToBackend)
		req.Headers = requestToBackend.Header

		resp, err := next(ctx, req)

		defer func() {
			externalSegment.End()
			closer(requestToBackend)
		}()

		return resp, err
	}
}

func toHttpRequest(req *proxy.Request) (*http.Request, func(r *http.Request), error) {
	requestToBackend, err := http.NewRequest(strings.ToTitle(req.Method), req.URL.String(), req.Body)
	if err != nil {
		return nil, func(r *http.Request) {

		}, err
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

	return requestToBackend, func(r *http.Request) {
		if r.Body != nil {
			r.Body.Close()
		}
	}, nil
}
