package metrics

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	type fields struct {
		segmentName string
		nextFactory proxy.FactoryFunc
		cfg         *config.EndpointConfig
	}
	type args struct {
		req *proxy.Request
		cfg *config.EndpointConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		app     *Application
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "given app and newrelic transaction is defined, it should start a transaction segment",
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(endpointConfig *config.EndpointConfig) (proxy.Proxy, error) {
					return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
						return nil, nil
					}, nil
				},
				cfg: &config.EndpointConfig{},
			},
			args: args{
				req: &proxy.Request{},
				cfg: &config.EndpointConfig{
					Endpoint: "/some-endpoint",
				},
			},
			app: func() *Application {
				return &Application{
					TransactionManager: func() TransactionManager {
						tx := NewMockTransaction(ctrl)
						tm := NewMockTransactionManager(ctrl)

						tm.EXPECT().TransactionFromContext(gomock.AssignableToTypeOf(context.Background())).
							Times(1).
							Return(tx)

						tx.EXPECT().StartSegment("(segment1) /some-endpoint").Return(&newrelic.Segment{}).Times(1)

						return tm
					}(),
					NRApplication: NewMockNRApplication(ctrl),
					Config:        Config{},
				}
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "given newrelic transaction is not defined, it should use the next proxy",
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(endpointConfig *config.EndpointConfig) (proxy.Proxy, error) {
					return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
						return nil, nil
					}, nil
				},
				cfg: &config.EndpointConfig{},
			},
			args: args{
				req: &proxy.Request{},
				cfg: &config.EndpointConfig{
					Endpoint: "/some-endpoint",
				},
			},
			app: &Application{
				TransactionManager: func() TransactionManager {
					tm := NewMockTransactionManager(ctrl)
					tm.EXPECT().TransactionFromContext(gomock.AssignableToTypeOf(context.Background())).
						Times(1).
						Return(nil)

					return tm
				}(),
				NRApplication: NewMockNRApplication(ctrl),
				Config:        Config{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "given app is not defined, it should use the next proxy",
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(endpointConfig *config.EndpointConfig) (proxy.Proxy, error) {
					return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
						return nil, nil
					}, nil
				},
				cfg: &config.EndpointConfig{
					Endpoint: "/some-endpoint",
				},
			},
			args: args{
				req: &proxy.Request{},
			},
			app:     nil,
			wantErr: assert.NoError,
		},
		{
			name: "given next factory returned error, it should return an error",
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(endpointConfig *config.EndpointConfig) (proxy.Proxy, error) {
					return nil, errors.New("errored")
				},
				cfg: &config.EndpointConfig{
					Endpoint: "/some-endpoint",
				},
			},
			args: args{
				req: &proxy.Request{},
			},
			app:     &Application{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				app = tt.app
				proxyFactory := ProxyFactory(tt.fields.segmentName, tt.fields.nextFactory)

				m, err := proxyFactory(tt.args.cfg)
				if !tt.wantErr(t, err, "ProxyFactory()") {
					return
				}

				_, _ = m(context.Background(), tt.args.req)
			},
		)
	}
}
