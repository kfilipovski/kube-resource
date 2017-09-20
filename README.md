# Kubernetes Custom Resources (CRD) Example

Example for building Kubernetes Custom Resources (CRD) extensions.

kube-resource demonstrates the CRD usage, it shows how to:

1. Connect to the Kubernetes cluster
2. Create the new CRD if it doesn't exist  
3. Create a controller that listens to events associated with new resources

## Organization

the example contain 4 files:

* pkg/spec/spec.go - define and register our example Resource class
* pkg/client/client.go - client library to create and use our Resource class (CRUD)
* pkg/controller/controller.go - controller to manage objects if our Resource class
* kube-resource.go - main part, demonstrate how to create, use, and watch our Resource objects
* resource-example.yaml - file to create example object of Resource class

## Running

```
# assumes you have a working kubeconfig, not required if operating in-cluster
go run *.go -kubeconf=$HOME/.kube/config
```


## Try

```
kubectl create -f resource-example.yaml
```

* The Metadata part contain standard Kubernetes properties like name, namespace, labels, and annotations
* The Spec contain the desired resource configuration
* The Status part is usually filled by the controller in response to Spec updates

## Credits

This example is based on:

* [yaronha/kube-crd](https://github.com/yaronha/kube-crd)
* [jbeda/tgik-controller](https://github.com/jbeda/tgik-controller)
