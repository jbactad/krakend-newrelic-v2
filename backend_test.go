package metrics

import (
	"context"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/stretchr/testify/assert"
)

func TestBackendFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	type fields struct {
		segmentName string
		nextFactory proxy.BackendFactory
		cfg         *config.Backend
	}
	type args struct {
		request *proxy.Request
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		app     *Application
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "given app is registered, it should start a new transaction segment",
			args: args{
				request: &proxy.Request{
					Method: "GET",
					URL: func() *url.URL {
						u, err := url.Parse("http://localhost:8080/some-endpoint")
						if err != nil {
							t.Fatal(err)
						}

						return u
					}(),
				},
			},
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(remote *config.Backend) proxy.Proxy {
					return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
						return &proxy.Response{}, nil
					}
				},
			},
			app: func() *Application {
				return &Application{
					TransactionManager: func() TransactionManager {
						seg := NewMockTransactionEnder(ctrl)
						tx := NewMockTransaction(ctrl)
						tp := NewMockTransactionManager(ctrl)

						tp.EXPECT().TransactionFromContext(gomock.AssignableToTypeOf(context.Background())).
							Times(1).
							Return(tx)

						tp.EXPECT().StartExternalSegment(tx, gomock.Any()).
							Times(1).
							Return(seg)

						seg.EXPECT().End().
							Times(1)

						return tp
					}(),
					NRApplication: NewMockNRApplication(ctrl),
					Config:        Config{},
				}
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "given app is not registered, it should not invoke backend proxy",
			args: args{
				request: &proxy.Request{},
			},
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(remote *config.Backend) proxy.Proxy {
					return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
						return &proxy.Response{}, nil
					}
				},
			},
			app:     nil,
			wantErr: assert.NoError,
		},
		{
			name: "given transaction is nil, it should not start a new transaction segment",
			args: args{
				request: &proxy.Request{},
			},
			fields: fields{
				segmentName: "segment1",
				nextFactory: func(remote *config.Backend) proxy.Proxy {
					return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
						return &proxy.Response{}, nil
					}
				},
			},
			app: func() *Application {
				return &Application{
					TransactionManager: func() TransactionManager {
						tp := NewMockTransactionManager(ctrl)

						tp.EXPECT().TransactionFromContext(gomock.AssignableToTypeOf(context.Background())).Times(1).
							Return(nil)

						return tp
					}(),
					NRApplication: NewMockNRApplication(ctrl),
					Config:        Config{},
				}
			}(),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				app = tt.app
				backendProxy := BackendFactory(tt.fields.segmentName, tt.fields.nextFactory)(tt.fields.cfg)
				_, err := backendProxy(context.Background(), tt.args.request)
				if !tt.wantErr(t, err, "backend()") {
					return
				}

			},
		)
	}
}
