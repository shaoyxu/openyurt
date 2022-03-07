/*
Copyright 2021 The OpenYurt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/openyurtio/openyurt/pkg/projectinfo"
)

var (
	namespace = "node"
	subsystem = strings.ReplaceAll(projectinfo.GetHubName(), "-", "_")
)

var (
	// Metrics provides access to all hub agent metrics.
	Metrics = newHubMetrics()
)

type HubMetrics struct {
	serversHealthyCollector   *prometheus.GaugeVec
	inFlightRequestsCollector *prometheus.GaugeVec
	inFlightRequestsGauge     prometheus.Gauge
	rejectedRequestsCounter   prometheus.Counter
	closableConnsCollector    *prometheus.GaugeVec
	proxyTrafficCollector     *prometheus.CounterVec
}

func newHubMetrics() *HubMetrics {
	serversHealthyCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "server_healthy_status",
			Help:      "healthy status of remote servers. 1: healthy, 0: unhealthy",
		},
		[]string{"server"})
	inFlightRequestsCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "in_flight_requests_collector",
			Help:      "collector of in flight requests handling by hub agent",
		},
		[]string{"verb", "resource", "subresources", "client"})
	inFlightRequestsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "in_flight_requests_total",
			Help:      "total of in flight requests handling by hub agent",
		})
	rejectedRequestsCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "rejected_requests_counter",
			Help:      "counter of rejected requests for exceeding in flight limit in hub agent",
		})
	closableConnsCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "closable_conns_collector",
			Help:      "collector of underlay tcp connection from hub agent to remote server",
		},
		[]string{"server"})
	proxyTrafficCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "proxy_traffic_collector",
			Help:      "collector of proxy response traffic by hub agent(unit: byte)",
		},
		[]string{"client", "verb", "resource", "subresources"})
	prometheus.MustRegister(serversHealthyCollector)
	prometheus.MustRegister(inFlightRequestsCollector)
	prometheus.MustRegister(inFlightRequestsGauge)
	prometheus.MustRegister(rejectedRequestsCounter)
	prometheus.MustRegister(closableConnsCollector)
	prometheus.MustRegister(proxyTrafficCollector)
	return &HubMetrics{
		serversHealthyCollector:   serversHealthyCollector,
		inFlightRequestsCollector: inFlightRequestsCollector,
		inFlightRequestsGauge:     inFlightRequestsGauge,
		rejectedRequestsCounter:   rejectedRequestsCounter,
		closableConnsCollector:    closableConnsCollector,
		proxyTrafficCollector:     proxyTrafficCollector,
	}
}

func (hm *HubMetrics) Reset() {
	hm.serversHealthyCollector.Reset()
	hm.inFlightRequestsCollector.Reset()
	hm.inFlightRequestsGauge.Set(float64(0))
	hm.closableConnsCollector.Reset()
	hm.proxyTrafficCollector.Reset()
}

func (hm *HubMetrics) ObserveServerHealthy(server string, status int) {
	hm.serversHealthyCollector.WithLabelValues(server).Set(float64(status))
}

func (hm *HubMetrics) IncInFlightRequests(verb, resource, subresource, client string) {
	hm.inFlightRequestsCollector.WithLabelValues(verb, resource, subresource, client).Inc()
	hm.inFlightRequestsGauge.Inc()
}

func (hm *HubMetrics) DecInFlightRequests(verb, resource, subresource, client string) {
	hm.inFlightRequestsCollector.WithLabelValues(verb, resource, subresource, client).Dec()
	hm.inFlightRequestsGauge.Dec()
}

func (hm *HubMetrics) IncRejectedRequestCounter() {
	hm.rejectedRequestsCounter.Inc()
}

func (hm *HubMetrics) IncClosableConns(server string) {
	hm.closableConnsCollector.WithLabelValues(server).Inc()
}

func (hm *HubMetrics) DecClosableConns(server string) {
	hm.closableConnsCollector.WithLabelValues(server).Dec()
}

func (hm *HubMetrics) SetClosableConns(server string, cnt int) {
	hm.closableConnsCollector.WithLabelValues(server).Set(float64(cnt))
}

func (hm *HubMetrics) AddProxyTrafficCollector(client, verb, resource, subresource string, size int) {
	if size > 0 {
		hm.proxyTrafficCollector.WithLabelValues(client, verb, resource, subresource).Add(float64(size))
	}
}
