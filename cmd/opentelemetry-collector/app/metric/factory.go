package metric

import (
	"context"
	"fmt"
	jMetrics "github.com/uber/jaeger-lib/metrics"
	"go.opentelemetry.io/otel/api/core"
	otelMetrics "go.opentelemetry.io/otel/api/metric"
	"go.uber.org/zap"
	"time"
)

// Factory implements metrics.Factory backed by OpenTelemetry.
type Factory struct {
	provider otelMetrics.Provider
	meter otelMetrics.Meter
	name string
	logger *zap.Logger
}

// New creates Factory.
func New(name string, provider otelMetrics.Provider, logger *zap.Logger) Factory {
	return Factory{
		name: name,
		provider: provider,
		meter: provider.Meter(name),
		logger: logger,
	}
}

func (o Factory) Counter(opts jMetrics.Options) jMetrics.Counter {
	kvs := toKeyValues(opts.Tags)
	c, err := o.meter.NewInt64Counter(opts.Name,
		otelMetrics.WithDescription(opts.Help),
		otelMetrics.WithKeys(toKeys(opts.Tags)...))
	if err != nil {
		o.logger.Fatal("could not create counter", zap.Error(err))
	}
	return &counter{wrapped: c, kvs: kvs}
}

func (o Factory) Gauge(opts jMetrics.Options) jMetrics.Gauge {
	kvs := toKeyValues(opts.Tags)
	m, err := o.meter.NewInt64Measure(opts.Name,
		otelMetrics.WithDescription(opts.Help),
		otelMetrics.WithKeys(toKeys(opts.Tags)...))
	if err != nil {
		o.logger.Fatal("could not create measure", zap.Error(err))
	}
	return int64gauge{wrapped: m, kvs: kvs}
}

func (o Factory) Timer(opts jMetrics.TimerOptions) jMetrics.Timer {
	kvs := toKeyValues(opts.Tags)
	m, err := o.meter.NewInt64Measure(opts.Name,
		otelMetrics.WithDescription(opts.Help),
		otelMetrics.WithKeys(toKeys(opts.Tags)...))
	if err != nil {
		o.logger.Fatal("could not create measure", zap.Error(err))
	}
	return int64gauge{wrapped: m, kvs: kvs}
}

func (o Factory) Histogram(opts jMetrics.HistogramOptions) jMetrics.Histogram {
	kvs := toKeyValues(opts.Tags)
	m, err := o.meter.NewFloat64Measure(opts.Name,
		otelMetrics.WithDescription(opts.Help),
		otelMetrics.WithKeys(toKeys(opts.Tags)...))
	if err != nil {
		o.logger.Fatal("could not create measure", zap.Error(err))
	}
	return float64gauge{wrapped: m, kvs: kvs}
}

func (o Factory) Namespace(opts jMetrics.NSOptions) jMetrics.Factory {
	otel := New(subName(o.name, opts.Name), o.provider, o.logger)
	return otel
}

func subName(a, b string) string {
	return fmt.Sprintf("%s_%s", a, b)
}

type counter struct {
	wrapped otelMetrics.Int64Counter
	kvs []core.KeyValue
}

func (c counter) Inc(v int64) {
	c.wrapped.Add(context.Background(), v, c.kvs...)
}

type int64gauge struct {
	wrapped otelMetrics.Int64Measure
	kvs []core.KeyValue
}

func (g int64gauge) Record(d time.Duration) {
	g.wrapped.Record(context.Background(), d.Nanoseconds(), g.kvs...)
}

func (g int64gauge) Update(v int64) {
	g.wrapped.Record(context.Background(), v, g.kvs...)
}

type float64gauge struct {
	wrapped otelMetrics.Float64Measure
	kvs []core.KeyValue
}

func (f float64gauge) Record(v float64) {
	f.wrapped.Record(context.Background(), v, f.kvs...)
}

func toKeys(tags map[string]string) []core.Key {
	var names []core.Key
	for k := range tags {
		names = append(names, core.Key(k))
	}
	return names
}

func toKeyValues(tags map[string]string) []core.KeyValue {
	var names []core.KeyValue
	for k, v := range tags {
		key := core.Key(k)
		names = append(names, key.String(v))
	}
	return names
}
