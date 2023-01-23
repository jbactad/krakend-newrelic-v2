package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Namespace for krakend_newrelic
const Namespace = "github_com/jbactad/krakend_newrelic_v2"

var (
	app *Application
)

// Config struct for NewRelic Krakend
type Config struct {
	InstrumentationRate int `json:"rate"`
}

type NRApplication interface {
	StartTransaction(name string, opts ...newrelic.TraceOption) *newrelic.Transaction
	RecordCustomEvent(eventType string, params map[string]interface{})
	RecordCustomMetric(name string, value float64)
	RecordLog(logEvent newrelic.LogData)
	WaitForConnection(timeout time.Duration) error
	Shutdown(timeout time.Duration)
}

type Transaction interface {
	End()
	SetName(name string)
	SetWebRequestHTTP(r *http.Request)
	SetWebRequest(r newrelic.WebRequest)
	SetWebResponse(w http.ResponseWriter) http.ResponseWriter
	StartSegmentNow() newrelic.SegmentStartTime
	StartSegment(name string) *newrelic.Segment
	InsertDistributedTraceHeaders(hdrs http.Header)
	NewGoroutine() *newrelic.Transaction
	GetTraceMetadata() newrelic.TraceMetadata
	GetLinkingMetadata() newrelic.LinkingMetadata
}

type TransactionEnder interface {
	End()
}

type TransactionManager interface {
	TransactionFromContext(ctx context.Context) Transaction
	StartExternalSegment(txn Transaction, request *http.Request) TransactionEnder
}

type Application struct {
	TransactionManager TransactionManager
	NRApplication
	Config Config
}

type NewRelicAppFactoryFunc func() (NRApplication, error)

// ConfigGetter gets config for NewRelic
func ConfigGetter(cfg config.ExtraConfig) (Config, error) {
	result := Config{}
	v, ok := cfg[Namespace]
	if !ok {
		return result, fmt.Errorf("namespace %s is not defined in extra_config", Namespace)
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("cannot map config to map string interface")
	}

	marshaledConf, err := json.Marshal(tmp)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(marshaledConf, &result)

	return result, err
}

// Register initializes the metrics collector.
// Returns a newrelic.Application instance.
func Register(
	ctx context.Context,
	cfg config.ExtraConfig,
	logger logging.Logger,
) *newrelic.Application {
	var err error
	app, err = NewApp(
		cfg, func() (NRApplication, error) {
			return newrelic.NewApplication(
				newrelic.ConfigFromEnvironment(),
				newrelic.ConfigAppLogDecoratingEnabled(true),
			)
		},
		newrelicWrapper{},
	)
	if err != nil {
		logger.Error("error initializing metrics collector", err.Error())
		return nil
	}

	err = app.WaitForConnection(time.Second * 5)
	if err != nil {
		logger.Error("error initializing metrics collector", err.Error())
	}

	return app.NRApplication.(*newrelic.Application)
}

// NewApp creates a new Application that wraps the newrelic application
func NewApp(
	cfg config.ExtraConfig,
	nrAppFactory NewRelicAppFactoryFunc,
	manager TransactionManager,
) (*Application, error) {
	conf, err := ConfigGetter(cfg)
	if err != nil {
		return nil, fmt.Errorf("no config for the NR module: %w", err)
	}

	nrApp, err := nrAppFactory()
	if err != nil {
		return nil, fmt.Errorf("unable to start the NR module: %w", err)
	}

	return &Application{manager, nrApp, conf}, nil
}

type newrelicWrapper struct {
}

func (t newrelicWrapper) StartExternalSegment(txn Transaction, request *http.Request) TransactionEnder {
	return newrelic.StartExternalSegment(txn.(*newrelic.Transaction), request)
}

func (t newrelicWrapper) TransactionFromContext(ctx context.Context) Transaction {
	return newrelic.FromContext(ctx)
}
