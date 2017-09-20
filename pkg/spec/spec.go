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

package spec

import (
	"reflect"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

const (
	// ResourceGroup fn group
	ResourceGroup string = "oracledx.com"
	// ResourceVersion fn version
	ResourceVersion string = "v1"
	// ResourcePlural resource plural name
	ResourcePlural string = "resources"
)

// Resource definition of Resource class
type Resource struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               ResourceSpec   `json:"spec"`
	Status             ResourceStatus `json:"status,omitempty"`
}

// ResourceSpec definition of Resource spec
type ResourceSpec struct {
	Foo string `json:"foo"`
	Bar bool   `json:"bar"`
	Baz int    `json:"baz,omitempty"`
}

// ResourceStatus definition of Resource status
type ResourceStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// ResourceList definition of Resources list
type ResourceList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Resource `json:"items"`
}

// CreateResource creates the Resource resource, ignore error if it already exists
func CreateResource(clientset apiextcs.Interface) error {
	resource := &apiextv1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{Name: ResourcePlural + "." + ResourceGroup},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group:   ResourceGroup,
			Version: ResourceVersion,
			Scope:   apiextv1beta1.NamespaceScoped,
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Plural: ResourcePlural,
				Kind:   reflect.TypeOf(Resource{}).Name(),
			},
		},
	}

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(resource)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err

	// TODO: add logic to wait for creation and exception handling
}

// SchemeGroupVersion creates a Rest client with the new Resource scheme
var SchemeGroupVersion = schema.GroupVersion{Group: ResourceGroup, Version: ResourceVersion}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Resource{},
		&ResourceList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func NewClientSet(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}
	return client, scheme, nil
}
