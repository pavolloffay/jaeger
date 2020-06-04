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

package grpcpluginexporter

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"

	"github.com/jaegertracing/jaeger/cmd/flags"
	jConfig "github.com/jaegertracing/jaeger/pkg/config"
	storageGrpc "github.com/jaegertracing/jaeger/plugin/storage/grpc"
)

func TestDefaultConfig(t *testing.T) {
	v, _ := jConfig.Viperize(DefaultOptions().AddFlags)
	opts := DefaultOptions()
	opts.InitFromViper(v)
	factory := &Factory{OptionsFactory: func() *storageGrpc.Options {
		return opts
	}}
	defaultCfg := factory.CreateDefaultConfig().(*Config)
	assert.Equal(t, "warn", defaultCfg.Configuration.PluginLogLevel)
}

func TestLoadConfigAndFlags(t *testing.T) {
	factories, err := config.ExampleComponents()
	require.NoError(t, err)

	v, c := jConfig.Viperize(DefaultOptions().AddFlags, flags.AddConfigFileFlag)
	err = c.ParseFlags([]string{"--grpc-storage-plugin.binary=/superstore", "--config-file=./testdata/jaeger-config.yaml"})
	require.NoError(t, err)

	err = flags.TryLoadConfigFile(v)
	require.NoError(t, err)

	factory := &Factory{OptionsFactory: func() *storageGrpc.Options {
		opts := DefaultOptions()
		opts.InitFromViper(v)
		require.Equal(t, "/superstore", opts.Configuration.PluginBinary)
		return opts
	}}

	factories.Exporters[TypeStr] = factory
	colConfig, err := config.LoadConfigFile(t, path.Join(".", "testdata", "config.yaml"), factories)
	require.NoError(t, err)
	require.NotNil(t, colConfig)

	cfg := colConfig.Exporters[TypeStr].(*Config)
	grpcCfg := cfg.Configuration
	assert.Equal(t, TypeStr, cfg.Name())
	assert.Equal(t, "/superstore", grpcCfg.PluginBinary)
	assert.Equal(t, "info", grpcCfg.PluginLogLevel)
	assert.Equal(t, "/doesnt/exist", grpcCfg.PluginConfigurationFile)
}
