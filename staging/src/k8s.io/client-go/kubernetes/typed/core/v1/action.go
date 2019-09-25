/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// ActionsGetter has a method to return a ActionInterface.
// A group's client should implement this interface.
type ActionsGetter interface {
	Actions(namespace string) ActionInterface
}

// ActionInterface has methods to work with Action resources.
type ActionInterface interface {
	Create(*v1.Action) (*v1.Action, error)
	Update(*v1.Action) (*v1.Action, error)
	UpdateStatus(*v1.Action) (*v1.Action, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Action, error)
	List(opts metav1.ListOptions) (*v1.ActionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Action, err error)
	ActionExpansion
}

// actions implements ActionInterface
type actions struct {
	client rest.Interface
	ns     string
}

// newActions returns a Actions
func newActions(c *CoreV1Client, namespace string) *actions {
	return &actions{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the action, and returns the corresponding action object, and an error if there is any.
func (c *actions) Get(name string, options metav1.GetOptions) (result *v1.Action, err error) {
	result = &v1.Action{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("actions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Actions that match those selectors.
func (c *actions) List(opts metav1.ListOptions) (result *v1.ActionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.ActionList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("actions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested actions.
func (c *actions) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("actions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a action and creates it.  Returns the server's representation of the action, and an error, if there is any.
func (c *actions) Create(action *v1.Action) (result *v1.Action, err error) {
	result = &v1.Action{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("actions").
		Body(action).
		Do().
		Into(result)
	return
}

// Update takes the representation of a action and updates it. Returns the server's representation of the action, and an error, if there is any.
func (c *actions) Update(action *v1.Action) (result *v1.Action, err error) {
	result = &v1.Action{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("actions").
		Name(action.Name).
		Body(action).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *actions) UpdateStatus(action *v1.Action) (result *v1.Action, err error) {
	result = &v1.Action{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("actions").
		Name(action.Name).
		SubResource("status").
		Body(action).
		Do().
		Into(result)
	return
}

// Delete takes name of the action and deletes it. Returns an error if one occurs.
func (c *actions) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("actions").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *actions) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("actions").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched action.
func (c *actions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Action, err error) {
	result = &v1.Action{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("actions").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
