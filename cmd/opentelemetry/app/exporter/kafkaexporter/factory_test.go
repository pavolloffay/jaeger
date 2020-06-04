// Copyright (c) 2020 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kafkaexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.uber.org/zap"

	jConfig "github.com/jaegertracing/jaeger/pkg/config"
	"github.com/jaegertracing/jaeger/plugin/storage/kafka"
)

func TestCreateTraceExporter(t *testing.T) {
	v, _ := jConfig.Viperize(DefaultOptions().AddFlags)
	opts := DefaultOptions()
	opts.InitFromViper(v)
	factory := &Factory{OptionsFactory: func() *kafka.Options {
		return opts
	}}
	exporter, err := factory.CreateTraceExporter(context.Background(), component.ExporterCreateParams{Logger: zap.NewNop()}, factory.CreateDefaultConfig())
	require.Nil(t, exporter)
	assert.Contains(t, err.Error(), "kafka: client has run out of available brokers to talk to (Is your cluster reachable?)")
}

func TestCreateTraceExporter_nilConfig(t *testing.T) {
	factory := &Factory{}
	exporter, err := factory.CreateTraceExporter(context.Background(), component.ExporterCreateParams{}, nil)
	require.Nil(t, exporter)
	assert.Contains(t, err.Error(), "could not cast configuration to jaeger_kafka")
}

func TestCreateMetricsExporter(t *testing.T) {
	f := Factory{OptionsFactory: DefaultOptions}
	mReceiver, err := f.CreateMetricsExporter(context.Background(), component.ExporterCreateParams{}, f.CreateDefaultConfig())
	assert.Equal(t, err, configerror.ErrDataTypeIsNotSupported)
	assert.Nil(t, mReceiver)
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := Factory{OptionsFactory: DefaultOptions}
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestType(t *testing.T) {
	factory := Factory{OptionsFactory: DefaultOptions}
	assert.Equal(t, configmodels.Type(TypeStr), factory.Type())
}
