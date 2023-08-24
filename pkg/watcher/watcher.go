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

package watcher

import (
	"strings"
	"sync"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	informerTimeout = time.Minute
	podResourceType = "pods"
	exporterLabel   = "export-labels"
)

type ObjListWatcher struct {
	k8sCli       *kubernetes.Clientset
	ResourceKind string
	informer     cache.SharedInformer
	stopChannel  chan struct{}

	LabelNames      *map[string]bool
	LabelPrefix     string
	ExposeAllLabels bool

	// PodMetrics holds all pod labels
	PodMetrics *map[string]map[string]string

	// Lock to syncronize the collector update with prometheus exporter
	Mx *sync.Mutex
}

func newK8sClient(kubeConfigPath string) *kubernetes.Clientset {
	var restConf *rest.Config
	var err error
	if kubeConfigPath == "" {
		// creates the in-cluster config
		restConf, err = rest.InClusterConfig()
		klog.Infoln("Using in cluster k8s config")
	} else {
		// use the current context in kubeconfig
		restConf, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		klog.Infoln("Using out cluster k8s config: ", kubeConfigPath)
	}
	if err != nil {
		klog.Infof("failed to get config: %v", err)
		return nil
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(restConf)
	if err != nil {
		klog.Fatalf("%v", err)
	}
	return clientset
}

func NewObjListWatcher(kubeConfigPath string) *ObjListWatcher {
	w := &ObjListWatcher{
		stopChannel:  make(chan struct{}),
		k8sCli:       newK8sClient(kubeConfigPath),
		ResourceKind: podResourceType,
	}
	if w.k8sCli == nil {
		return w
	}

	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = "" // do not filter events
	}
	objListWatcher := cache.NewFilteredListWatchFromClient(
		w.k8sCli.CoreV1().RESTClient(),
		w.ResourceKind,
		metav1.NamespaceAll,
		optionsModifier,
	)

	w.informer = cache.NewSharedInformer(objListWatcher, nil, 0)
	w.stopChannel = make(chan struct{})
	w.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			w.handleAdded(obj)
		},
		DeleteFunc: func(obj interface{}) {
			w.handleDeleted(obj)
		},
	})
	return w
}

func (w *ObjListWatcher) Run() {
	go w.informer.Run(w.stopChannel)
	timeoutCh := make(chan struct{})
	timeoutTimer := time.AfterFunc(informerTimeout, func() {
		close(timeoutCh)
	})
	defer timeoutTimer.Stop()
	if !cache.WaitForCacheSync(timeoutCh, w.informer.HasSynced) {
		klog.Fatalf("watcher timed out waiting for caches to sync")
	}
}

func (w *ObjListWatcher) Stop() {
	close(w.stopChannel)
}

func (w *ObjListWatcher) handleAdded(obj interface{}) {
	switch w.ResourceKind {
	case podResourceType:
		pod, ok := obj.(*k8sv1.Pod)
		if !ok {
			klog.Infof("Could not convert obj: %v", w.ResourceKind)
			return
		}
		if _, ok := pod.ObjectMeta.Labels[exporterLabel]; !ok && !w.ExposeAllLabels {
			klog.V(5).Infof("Pod %s/%s does not have the label to enable label exporting", pod.Namespace, pod.Name)
			return
		}
		if _, ok := (*w.PodMetrics)[pod.Namespace+pod.Name]; !ok {
			(*w.PodMetrics)[pod.Namespace+"/"+pod.Name] = map[string]string{}
		}
		w.Mx.Lock()
		for label := range pod.ObjectMeta.Labels {
			if strings.Contains(label, w.LabelPrefix) || w.ExposeAllLabels {
				(*w.PodMetrics)[pod.Namespace+"/"+pod.Name][label] = pod.ObjectMeta.Labels[label]
				(*w.LabelNames)[label] = true
			}
		}
		w.Mx.Unlock()

	default:
		klog.Infof("Watcher does not support object type %s", w.ResourceKind)
		return
	}
}

// TODO: ranging the labels of all pods can become expensive at scale
func (w *ObjListWatcher) handleDeleted(obj interface{}) {
	switch w.ResourceKind {
	case podResourceType:
		pod, ok := obj.(*k8sv1.Pod)
		if !ok {
			klog.Infof("Could not convert obj: %v", w.ResourceKind)
			return
		}
		w.Mx.Lock()
		podLabels := pod.ObjectMeta.Labels
		delete((*w.PodMetrics), pod.Namespace+"/"+pod.Name)

	out:
		for podLabel := range podLabels {
			for pod := range *w.PodMetrics {
				if _, ok := (*w.PodMetrics)[pod][podLabel]; ok {
					// this label is used by another pod, so we cannot remove it
					break out
				}
			}
			delete((*w.LabelNames), podLabel)
		}

		w.Mx.Unlock()

	default:
		klog.Infof("Watcher does not support object type %s", w.ResourceKind)
		return
	}
}
