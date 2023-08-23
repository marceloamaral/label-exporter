/*
Copyright 2023.

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

package exporter

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusCollector struct {
	LabelNames *map[string]bool

	// PodMetrics holds all pod labels
	PodMetrics *map[string]map[string]string

	// Lock to syncronize the collector update with prometheus exporter
	Mx sync.Mutex
}

// NewPrometheusExporter create and initialize all the PrometheusCollector structures
func NewPrometheusExporter() *PrometheusCollector {
	exporter := PrometheusCollector{}
	return &exporter
}

func (p *PrometheusCollector) GetDescription() (*prometheus.Desc, []string) {
	orderedLabels := p.getLabels()
	return prometheus.NewDesc(
		prometheus.BuildFQName("", "label", "exporter"),
		"Labeled pod labels",
		append([]string{"pod_name", "pod_namespace"}, orderedLabels...), nil,
	), orderedLabels
}

// Describe implements the prometheus.Collector interface
func (p *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	desc, _ := p.GetDescription()
	ch <- desc
}

// Collect implements the prometheus.Collector interface
func (p *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	p.Mx.Lock()
	defer p.Mx.Unlock()

	// get new metric description, since labels might have changed
	desc, orderedLabels := p.GetDescription()

	for key := range *p.PodMetrics {
		split := strings.Split(key, "/")
		podNameSpace := split[0]
		podName := split[1]

		podLabels := []string{podName, podNameSpace}
		for _, label := range orderedLabels {
			value := (*p.PodMetrics)[key][label]
			podLabels = append(podLabels, value)
		}

		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.CounterValue,
			0.0,
			podLabels...,
		)
	}
}

func (p *PrometheusCollector) getLabels() []string {
	keys := []string{}
	if p.LabelNames == nil {
		return keys
	}
	for key := range *p.LabelNames {
		keys = append(keys, key)
	}
	return keys
}
