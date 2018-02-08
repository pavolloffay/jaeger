package local

import (
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

// Factory implements storage.Factory and creates storage components backed by memory store.
type Factory struct {
	metricsFactory metrics.Factory
	logger         *zap.Logger
	storage *Storage
}

func (f *Factory) Initialize(metricsFactory metrics.Factory, logger *zap.Logger) error {
	s := &Storage{}
	return s.Start()
}

func (f *Factory) CreateSpanReader() (spanstore.Reader, error) {
	return nil, nil
}

func (f *Factory) CreateSpanWriter() (spanstore.Writer, error) {
	return f.storage, nil
}

func (f *Factory) CreateDependencyReader() (dependencystore.Reader, error) {
	return nil, nil
}
