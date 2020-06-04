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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
	"go.opentelemetry.io/collector/service/builder"

	jflags "github.com/jaegertracing/jaeger/cmd/flags"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/defaults"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/receiver/kafkareceiver"
	jconfig "github.com/jaegertracing/jaeger/pkg/config"
	"github.com/jaegertracing/jaeger/pkg/version"
	"github.com/jaegertracing/jaeger/plugin/storage"
)

func main() {
	handleErr := func(err error) {
		if err != nil {
			log.Fatalf("Failed to run the service: %v", err)
		}
	}

	ver := version.Get()
	info := service.ApplicationStartInfo{
		ExeName:  "jaeger-opentelemetry-ingester",
		LongName: "Jaeger OpenTelemetry Ingester",
		Version:  ver.GitVersion,
		GitHash:  ver.GitCommit,
	}

	v := viper.New()

	storageType := os.Getenv(storage.SpanStorageTypeEnvVar)
	if storageType == "" {
		storageType = "cassandra"
	}

	cmpts := defaults.Components(v)
	cfgFactory := func(otelViper *viper.Viper, f config.Factories) (*configmodels.Config, error) {
		cfgConfig := defaults.ComponentSettings{
			ComponentType: defaults.Ingester,
			Factories:     cmpts,
			StorageType:   storageType,
		}
		cfg, err := cfgConfig.CreateDefaultConfig()
		if err != nil {
			return nil, err
		}
		if len(builder.GetConfigFile()) > 0 {
			otelCfg, err := service.FileLoaderConfigFactory(otelViper, f)
			if err != nil {
				return nil, err
			}
			err = defaults.MergeConfigs(cfg, otelCfg)
			if err != nil {
				return nil, err
			}
		}
		return cfg, nil
	}

	svc, err := service.New(service.Parameters{
		ApplicationStartInfo: info,
		Factories:            cmpts,
		ConfigFactory:        cfgFactory,
	})
	handleErr(err)

	// Add Jaeger specific flags to service command
	// this passes flag values to viper.
	storageFlags, err := app.AddStorageFlags(storageType, false)
	if err != nil {
		handleErr(err)
	}

	cmd := svc.Command()
	jconfig.AddFlags(v,
		cmd,
		kafkareceiver.AddFlags,
		app.AddComponentFlags,
		storageFlags,
	)

	// parse flags to propagate Jaeger config file flag value to viper
	cmd.ParseFlags(os.Args)
	err = jflags.TryLoadConfigFile(v)
	if err != nil {
		handleErr(fmt.Errorf("could not load Jaeger configuration file %w", err))
	}

	err = svc.Start()
	handleErr(err)
}
