// Copyright (c) 2018 The Jaeger Authors.
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

package grpc

import (
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"

	"github.com/jaegertracing/jaeger/cmd/agent/app/httpserver"
	aReporter "github.com/jaegertracing/jaeger/cmd/agent/app/reporter"
)

// ProxyBuilder holds objects communicating with collector
type ProxyBuilder struct {
	reporter aReporter.Reporter
	manager  httpserver.ClientConfigManager
}

// NewCollectorProxy creates ProxyBuilder
func NewCollectorProxy(o *Options, mFactory metrics.Factory, logger *zap.Logger) *ProxyBuilder {
	// It does not return error if the collector is not running
	// a way to fail immediately is to call WithBlock and WithTimeout
	var conn *grpc.ClientConn
	if len(o.CollectorHostPort) > 1 {
		r, _ := manual.GenerateAndRegisterManualResolver()
		var resolvedAddrs []resolver.Address
		for _, addr := range o.CollectorHostPort {
			resolvedAddrs = append(resolvedAddrs, resolver.Address{Addr: addr})
		}
		r.InitialAddrs(resolvedAddrs)
		conn, _ = grpc.Dial(r.Scheme()+":///round_robin", grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	} else {
		conn, _ = grpc.Dial(o.CollectorHostPort[0], grpc.WithInsecure())
	}
	return &ProxyBuilder{
		reporter: aReporter.WrapWithMetrics(NewReporter(conn, logger), "grpc-reporter", mFactory),
		manager:  httpserver.WrapWithMetrics(NewSamplingManager(conn), mFactory)}
}

// GetReporter returns Reporter
func (b ProxyBuilder) GetReporter() aReporter.Reporter {
	return b.reporter
}

// GetManager returns manager
func (b ProxyBuilder) GetManager() httpserver.ClientConfigManager {
	return b.manager
}
