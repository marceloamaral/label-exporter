/*
Copyright 2021.

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

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/marceloamaral/label-exporter/pkg/exporter"
	"github.com/marceloamaral/label-exporter/pkg/watcher"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	"k8s.io/klog/v2"
)

var (
	address         = flag.String("address", "0.0.0.0:9102", "bind address")
	kubeconfig      = flag.String("kubeconfig", "", "absolute path to the kubeconfig file, if empty we use the in-cluster configuration")
	metricsPath     = flag.String("metrics-path", "/metrics", "metrics path")
	labelPrefix     = flag.String("label-prefix", "le__", "only labels with this prefix will be exported to minimize the prometheus metric cardinality")
	exposeAllLabels = flag.Bool("expose-all", false, "expose all labels, if true the label-prefix will be ignored and all labels of all pods will be exporterd")

	// PrometheusCollector implements the external Collector interface provided by the Prometheus client
	PrometheusCollector *exporter.PrometheusCollector
	// Watcher register in the kubernetes apiserver to watch for pod events to add or remove it from the ContainersMetrics map
	Watcher *watcher.ObjListWatcher
)

func main() {
	labelNames := map[string]bool{}
	podMetrics := map[string]map[string]string{}

	PrometheusCollector = exporter.NewPrometheusExporter()
	PrometheusCollector.LabelNames = &labelNames
	PrometheusCollector.PodMetrics = &podMetrics

	Watcher := watcher.NewObjListWatcher(*kubeconfig)
	Watcher.Mx = &PrometheusCollector.Mx
	Watcher.LabelPrefix = *labelPrefix
	Watcher.ExposeAllLabels = *exposeAllLabels
	Watcher.LabelNames = &labelNames
	Watcher.PodMetrics = &podMetrics
	Watcher.Run()

	prometheus.MustRegister(version.NewCollector("label_exporter"))
	prometheus.MustRegister(PrometheusCollector)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
                        <head><title>Pod Labels Exporter</title></head>
                        <body>
                        <h1>Pod Labels Exporter</h1>
                        <p><a href="` + *metricsPath + `">Metrics</a></p>
                        </body>
                        </html>`))
		if err != nil {
			klog.Fatalf("%s", fmt.Sprintf("failed to write response: %v", err))
		}
	})
	klog.Fatalf("%v", http.ListenAndServe(*address, nil))
}
