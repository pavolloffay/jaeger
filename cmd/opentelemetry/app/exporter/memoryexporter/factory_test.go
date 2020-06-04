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

package memoryexporter

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.uber.org/zap"

	jConfig "github.com/jaegertracing/jaeger/pkg/config"
)

func cleanup() {
	instance = nil
}

func TestCreateTraceExporter(t *testing.T) {
	defer cleanup()

	v, _ := jConfig.Viperize(AddFlags)
	factory := NewFactory(v)
	exporter, err := factory.CreateTraceExporter(context.Background(), component.ExporterCreateParams{Logger: zap.NewNop()}, factory.CreateDefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func TestCreateTraceExporter_nilConfig(t *testing.T) {
	defer cleanup()

	factory := &Factory{}
	exporter, err := factory.CreateTraceExporter(context.Background(), component.ExporterCreateParams{}, nil)
	require.Nil(t, exporter)
	assert.Contains(t, err.Error(), "could not cast configuration to jaeger_memory")
}

func TestCreateMetricsExporter(t *testing.T) {
	defer cleanup()

	f := NewFactory(viper.New())
	mReceiver, err := f.CreateMetricsExporter(context.Background(), component.ExporterCreateParams{}, f.CreateDefaultConfig())
	assert.Equal(t, err, configerror.ErrDataTypeIsNotSupported)
	assert.Nil(t, mReceiver)
}

func TestCreateDefaultConfig(t *testing.T) {
	defer cleanup()

	factory := NewFactory(viper.New())
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestType(t *testing.T) {
	defer cleanup()

	factory := Factory{}
	assert.Equal(t, configmodels.Type(TypeStr), factory.Type())
}

func TestSingleton(t *testing.T) {
	defer cleanup()

	f := NewFactory(viper.New())
	logger := zap.NewNop()
	assert.Nil(t, instance)
	exp, err := f.CreateTraceExporter(context.Background(), component.ExporterCreateParams{Logger: logger}, &Config{})
	require.NoError(t, err)
	require.NotNil(t, exp)
	previousInstance := instance
	exp, err = f.CreateTraceExporter(context.Background(), component.ExporterCreateParams{Logger: logger}, &Config{})
	require.NoError(t, err)
	require.NotNil(t, exp)
	assert.Equal(t, previousInstance, instance)
}
