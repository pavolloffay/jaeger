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
	"io"
	"log"
	"os"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	jaegerClientConfig "github.com/uber/jaeger-client-go/config"
	jaegerClientZapLog "github.com/uber/jaeger-client-go/log/zap"
	"github.com/uber/jaeger-lib/metrics"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
	"go.opentelemetry.io/collector/service/builder"
	"go.uber.org/zap"

	collectorApp "github.com/jaegertracing/jaeger/cmd/collector/app"
	jflags "github.com/jaegertracing/jaeger/cmd/flags"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/defaults"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/exporter/badgerexporter"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/exporter/cassandraexporter"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/exporter/elasticsearchexporter"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/exporter/grpcpluginexporter"
	"github.com/jaegertracing/jaeger/cmd/opentelemetry/app/exporter/memoryexporter"
	queryApp "github.com/jaegertracing/jaeger/cmd/query/app"
	"github.com/jaegertracing/jaeger/cmd/query/app/querysvc"
	jConfig "github.com/jaegertracing/jaeger/pkg/config"
	"github.com/jaegertracing/jaeger/pkg/version"
	pluginStorage "github.com/jaegertracing/jaeger/plugin/storage"
	cassandraStorage "github.com/jaegertracing/jaeger/plugin/storage/cassandra"
	esStorage "github.com/jaegertracing/jaeger/plugin/storage/es"
	grpcStorage "github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/storage"
)

func main() {
	handleErr := func(err error) {
		if err != nil {
			log.Fatalf("Failed to run the service: %v", err)
		}
	}

	ver := version.Get()
	info := service.ApplicationStartInfo{
		ExeName:  "jaeger-opentelemetry-all-in-one",
		LongName: "Jaeger OpenTelemetry all-in-one",
		Version:  ver.GitVersion,
		GitHash:  ver.GitCommit,
	}

	v := viper.New()
	storageType := os.Getenv(pluginStorage.SpanStorageTypeEnvVar)
	if storageType == "" {
		storageType = "memory"
	}

	cmpts := defaults.Components(v)
	cfgFactory := func(otelViper *viper.Viper, f config.Factories) (*configmodels.Config, error) {
		collectorOpts := &collectorApp.CollectorOptions{}
		collectorOpts.InitFromViper(v)
		cfgOpts := defaults.ComponentSettings{
			ComponentType:  defaults.AllInOne,
			Factories:      cmpts,
			StorageType:    storageType,
			ZipkinHostPort: collectorOpts.CollectorZipkinHTTPHostPort,
		}
		cfg, err := cfgOpts.CreateDefaultConfig()
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
	storageFlags, err := app.AddStorageFlags(storageType, true)
	if err != nil {
		handleErr(err)
	}
	cmd := svc.Command()
	jConfig.AddFlags(v,
		cmd,
		app.AddComponentFlags,
		storageFlags,
		queryApp.AddFlags,
	)

	// parse flags to propagate Jaeger config file flag value to viper
	cmd.ParseFlags(os.Args)
	err = jflags.TryLoadConfigFile(v)
	if err != nil {
		handleErr(fmt.Errorf("could not load Jaeger configuration file %w", err))
	}

	go func() {
		err = svc.Start()
		handleErr(err)
	}()

	for state := range svc.GetStateChannel() {
		if state == service.Running {
			break
		}
	}
	exp := getStorageExporter(storageType, svc.GetExporters()[configmodels.TracesDataType])
	if exp == nil {
		svc.ReportFatalError(fmt.Errorf("exporter type for storage %s not found", storageType))
	}
	queryServer, tracerCloser, err := startQuery(v, svc.GetLogger(), exp)
	if err != nil {
		svc.ReportFatalError(err)
	}
	for state := range svc.GetStateChannel() {
		if state == service.Closing {
			if queryServer != nil {
				queryServer.Close()
			}
			if tracerCloser != nil {
				tracerCloser.Close()
			}
		} else if state == service.Closed {
			break
		}
	}
}

// getStorageExporter returns exporter for given storage type
// The storage type can contain a comma separated list of storage types
// the function does not handle this as the all-in-one should be used for a simple deployments with a single storage.
func getStorageExporter(storageType string, exporters map[configmodels.Exporter]component.Exporter) configmodels.Exporter {
	// replace `-` to `_` because grpc-plugin exporter is named as jaeger_grpc_plugin
	storageExporter := fmt.Sprintf("jaeger_%s", strings.Replace(storageType, "-", "_", -1))
	for k := range exporters {
		if storageExporter == k.Name() {
			return k
		}
	}
	return nil
}

func startQuery(v *viper.Viper, logger *zap.Logger, exporter configmodels.Exporter) (*queryApp.Server, io.Closer, error) {
	storageFactory, err := getFactory(exporter, v, logger)
	if err != nil {
		return nil, nil, err
	}
	spanReader, err := storageFactory.CreateSpanReader()
	if err != nil {
		return nil, nil, err
	}
	dependencyReader, err := storageFactory.CreateDependencyReader()
	if err != nil {
		return nil, nil, err
	}
	queryOpts := new(queryApp.QueryOptions).InitFromViper(v, logger)
	queryServiceOptions := queryOpts.BuildQueryServiceOptions(storageFactory, logger)
	queryService := querysvc.NewQueryService(
		spanReader,
		dependencyReader,
		*queryServiceOptions)

	tracerCloser := initTracer(logger)
	server := queryApp.NewServer(logger, queryService, queryOpts, opentracing.GlobalTracer())
	if err := server.Start(); err != nil {
		return nil, nil, err
	}
	return server, tracerCloser, nil
}

func getFactory(exporter configmodels.Exporter, v *viper.Viper, logger *zap.Logger) (storage.Factory, error) {
	switch exporter.Name() {
	case "jaeger_elasticsearch":
		archiveOpts := esStorage.NewOptions("es-archive")
		archiveOpts.InitFromViper(v)
		primaryConfig := exporter.(*elasticsearchexporter.Config)
		f := esStorage.NewFactory()
		f.InitFromOptions(*esStorage.NewOptionsFromConfig(primaryConfig.Primary.Configuration, archiveOpts.Primary.Configuration))
		if err := f.Initialize(metrics.NullFactory, logger); err != nil {
			return nil, err
		}
		return f, nil
	case "jaeger_cassandra":
		archiveOpts := cassandraStorage.NewOptions("cassandra-archive")
		archiveOpts.InitFromViper(v)
		primaryConfig := exporter.(*cassandraexporter.Config)
		f := cassandraStorage.NewFactory()
		f.InitFromOptions(cassandraStorage.NewOptionsFromConfig(primaryConfig.Primary.Configuration, archiveOpts.Primary.Configuration))
		if err := f.Initialize(metrics.NullFactory, logger); err != nil {
			return nil, err
		}
		return f, nil
	case "jaeger_grpc_plugin":
		primaryConfig := exporter.(*grpcpluginexporter.Config)
		f := grpcStorage.NewFactory()
		f.InitFromOptions(primaryConfig.Options)
		if err := f.Initialize(metrics.NullFactory, logger); err != nil {
			return nil, err
		}
		return f, nil
	case "jaeger_memory":
		return memoryexporter.GetFactory(), nil
	case "jaeger_badger":
		return badgerexporter.GetFactory(), nil
	default:
		return nil, fmt.Errorf("storage type %s cannot be used with all-in-one", exporter.Name())
	}
}

func initTracer(logger *zap.Logger) io.Closer {
	traceCfg := &jaegerClientConfig.Configuration{
		ServiceName: "jaeger-query",
		Sampler: &jaegerClientConfig.SamplerConfig{
			Type:  "const",
			Param: 1.0,
		},
		RPCMetrics: true,
	}
	traceCfg, err := traceCfg.FromEnv()
	if err != nil {
		logger.Fatal("Failed to read tracer configuration", zap.Error(err))
	}
	tracer, closer, err := traceCfg.NewTracer(
		jaegerClientConfig.Logger(jaegerClientZapLog.NewLogger(logger)),
	)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	opentracing.SetGlobalTracer(tracer)
	return closer
}
