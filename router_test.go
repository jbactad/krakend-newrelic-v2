package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		req *http.Request
	}
	callCount := 0
	ginMiddlewareProvider = func(application NRApplication) gin.HandlerFunc {
		return func(context *gin.Context) {
			callCount++
		}
	}
	tests := []struct {
		name string
		args args
		want int
		app  *Application
	}{
		{
			name: "given app is defined, it should create new NewRelic middleware",
			args: args{
				req: func() *http.Request {
					req, err := http.NewRequestWithContext(
						context.Background(),
						http.MethodGet,
						"http://localhost/somewhere",
						nil,
					)
					if err != nil {
						t.Fatal(err)
					}

					return req
				}(),
			},
			want: 1,
			app: func() *Application {
				return &Application{
					NRApplication: NewMockNRApplication(ctrl),
					Config: Config{
						InstrumentationRate: 100,
					},
				}
			}(),
		},
		{
			name: "given threshold 0, it should not invoke newrelic middleware",
			args: args{
				req: func() *http.Request {
					req, err := http.NewRequestWithContext(
						context.Background(),
						http.MethodGet,
						"http://localhost/somewhere",
						nil,
					)
					if err != nil {
						t.Fatal(err)
					}

					return req
				}(),
			},
			want: 0,
			app: func() *Application {
				return &Application{
					NRApplication: NewMockNRApplication(ctrl),
					Config: Config{
						InstrumentationRate: 0,
					},
				}
			}(),
		},
		{
			name: "given app is nil, it should return an err",
			args: args{
				req: func() *http.Request {
					req, err := http.NewRequestWithContext(
						context.Background(),
						http.MethodGet,
						"http://localhost/somewhere",
						nil,
					)
					if err != nil {
						t.Fatal(err)
					}

					return req
				}(),
			},
			want: 0,
			app:  nil,
		},
		{
			name: "given rate is 99, it should invoke newrelic middleware",
			args: args{
				req: func() *http.Request {
					req, err := http.NewRequestWithContext(
						context.Background(),
						http.MethodGet,
						"http://localhost/somewhere",
						nil,
					)
					if err != nil {
						t.Fatal(err)
					}

					return req
				}(),
			},
			want: 1,
			app: func() *Application {
				return &Application{
					NRApplication: NewMockNRApplication(ctrl),
					Config: Config{
						InstrumentationRate: 99,
					},
				}
			}(),
		},
		{
			name: "given rate is 1, it should invoke newrelic middleware",
			args: args{
				req: func() *http.Request {
					req, err := http.NewRequestWithContext(
						context.Background(),
						http.MethodGet,
						"http://localhost/somewhere",
						nil,
					)
					if err != nil {
						t.Fatal(err)
					}

					return req
				}(),
			},
			want: 0,
			app: func() *Application {
				return &Application{
					NRApplication: NewMockNRApplication(ctrl),
					Config: Config{
						InstrumentationRate: 1,
					},
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				app = tt.app
				callCount = 0

				w := httptest.NewRecorder()
				gin.SetMode(gin.TestMode)
				_, e := gin.CreateTestContext(w)

				h := Middleware()
				e.Use(h)

				e.ServeHTTP(w, tt.args.req)
				assert.EqualValues(t, tt.want, callCount)
			},
		)
	}
}

func TestHandlerFactory(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		handlerFactory router.HandlerFactory
		cfg            *config.EndpointConfig
	}

	tests := []struct {
		name string
		args args
		app  *Application
	}{
		{
			name: "given app is not nil, it should call NewGoroutine from transaction",
			args: args{
				cfg: &config.EndpointConfig{
					Method:   "GET",
					Endpoint: "some-endpoint",
				},
				handlerFactory: func(config *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
					return func(c *gin.Context) {
					}
				},
			},
			app: &Application{
				TransactionManager: func() TransactionManager {
					tm := NewMockTransactionManager(ctrl)
					tx := NewMockTransaction(ctrl)
					tm.EXPECT().TransactionFromContext(gomock.AssignableToTypeOf(&gin.Context{})).
						Times(1).
						Return(tx)
					tx.EXPECT().SetName(gomock.AssignableToTypeOf("")).
						Times(1)

					return tm
				}(),
				NRApplication: NewMockNRApplication(ctrl),
				Config:        Config{},
			},
		},
		{
			name: "given app is nil, it should not set transaction name",
			args: args{
				cfg: &config.EndpointConfig{
					Endpoint: "some-endpoint",
				},
				handlerFactory: func(config *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
					return func(c *gin.Context) {
					}
				},
			},
			app: nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				app = tt.app
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				p := func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
					return nil, nil
				}
				HandlerFactory(tt.args.handlerFactory)(tt.args.cfg, p)(c)
			},
		)
	}
}
