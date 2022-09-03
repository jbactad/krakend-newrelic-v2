//go:generate mockgen -source=metrics_collector.go -destination=newrelic_mocks.go -package=metrics
package metrics

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luraproject/lura/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	type args struct {
		cfg                config.ExtraConfig
		nrFactory          NewRelicAppFactoryFunc
		transactionManager TransactionManager
	}
	tests := []struct {
		name    string
		args    args
		want    *Application
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "given config is valid, it should return an app instance",
			args: args{
				cfg: map[string]interface{}{
					Namespace: map[string]interface{}{
						"rate": 0,
					},
				},
				nrFactory: func() (NRApplication, error) {
					return NewMockNRApplication(ctrl), nil
				},
				transactionManager: NewMockTransactionManager(ctrl),
			},
			want: &Application{
				TransactionManager: NewMockTransactionManager(ctrl),
				NRApplication:      NewMockNRApplication(ctrl),
				Config: Config{
					InstrumentationRate: 0,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "given error while creating newrelic app, it should return the error",
			args: args{
				cfg: map[string]interface{}{
					Namespace: map[string]interface{}{
						"rate": 0,
					},
				},
				nrFactory: func() (NRApplication, error) {
					return nil, errors.New("unable to create newrelic app")
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "given invalid config, it should return the error",
			args: args{
				cfg: map[string]interface{}{
					"invalid/namespace": map[string]interface{}{
						"rate": 0,
					},
				},
				nrFactory: func() (NRApplication, error) {
					return NewMockNRApplication(ctrl), nil
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "given invalid config, it should return an error",
			args: args{
				cfg: map[string]interface{}{
					Namespace: nil,
				},
				nrFactory: func() (NRApplication, error) {
					return NewMockNRApplication(ctrl), nil
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "given json marshaller error, it should return an error",
			args: args{
				cfg: map[string]interface{}{
					Namespace: map[string]interface{}{
						"rate": func() {},
					},
				},
				nrFactory: func() (NRApplication, error) {
					return NewMockNRApplication(ctrl), nil
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewApp(tt.args.cfg, tt.args.nrFactory, tt.args.transactionManager)
				if !tt.wantErr(
					t, err,
					"NewApp(%v, %p, %v)", tt.args.cfg, tt.args.nrFactory, tt.args.transactionManager,
				) {
					return
				}
				assert.Equalf(t, tt.want, got, "NewApp(%v, %p)", tt.args.cfg, tt.args.nrFactory)
			},
		)
	}
}
