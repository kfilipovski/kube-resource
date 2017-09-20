/*
Copyright 2017 Kire Filipovski

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
	"time"

	"github.com/kfilipovski/kube-resource/pkg/controller"
	"github.com/kfilipovski/kube-resource/pkg/spec"

	"flag"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {

	kubeconf := flag.String("kubeconf", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := GetClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// create clientset and create our Resource definition, this only need to run once
	clientset, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// NOTE: if the Resource definition exist our CreateResource function is set to exit without an error
	err = spec.CreateResource(clientset)
	if err != nil {
		panic(err)
	}

	// Wait for the Resource definition to be created before we use it (only needed if its a new one)
	time.Sleep(3 * time.Second)

	// Start Resource Controller

	c := controller.New(config)
	stopChan := make(chan struct{})
	defer close(stopChan)

	go c.Run(stopChan)

	// Wait forever
	select {}
}
