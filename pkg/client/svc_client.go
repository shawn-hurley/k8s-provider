package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/konveyor/analyzer-lsp/provider"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type k8sResourcePathConditionInfo struct {
	path     string
	resource string
}

type k8sServericeClient struct {
	config provider.InitConfig
	client *dynamic.DynamicClient
}

// Evaluate implements provider.ServiceClient.
func (c *k8sServericeClient) Evaluate(ctx context.Context, cap string, conditionInfo []byte) (provider.ProviderEvaluateResponse, error) {
	if cap == "k8s-resource-path" {
		a := &k8sResourcePathConditionInfo{}
		err := yaml.Unmarshal(conditionInfo, a)
		if err != nil {
			panic(err)
		}
		// gvr for resource, in future make this better for the user not just a space seperated thing.
		// We could also use GVK and look up GVR which might be more user friendly
		gvrStringParts := strings.Split(a.resource, " ")
		gvr := schema.GroupVersionResource{
			Group:    gvrStringParts[0],
			Resource: gvrStringParts[2],
			Version:  gvrStringParts[1],
		}

		resourceClientInterface := c.client.Resource(gvr)
		objs, err := resourceClientInterface.List(context.Background(), v1.ListOptions{})
		if err != nil {
			panic(err)
		}
		unstrs := []unstructured.Unstructured{}
		for _, obj := range objs.Items {
			// Here we need to get each thing, and ask it about the path that we got.
			if s, ok := obj.Object["spec"]; ok {
				spec, ok := s.(map[string]interface{})
				if !ok {
					continue
				}
				if t, ok := spec["template"]; ok {
					template, ok := t.(map[string]interface{})
					if !ok {
						continue
					}
					if s, ok := template["spec"]; ok {
						spec, ok := s.(map[string]interface{})
						if !ok {
							continue
						}
						if c, ok := spec["containers"]; ok {
							containers, ok := c.([]interface{})
							if !ok {
								continue
							}
							for _, co := range containers {
								if container, ok := co.(map[string]interface{}); ok {
									if _, ok := container["livenessProbe"]; !ok {
										unstrs = append(unstrs, obj)
									}
								}
							}
						}
					}
				}

			}
		}
		fmt.Printf("%#v", unstrs)
	}
	return provider.ProviderEvaluateResponse{}, nil
}

// GetDependencies implements provider.ServiceClient.
func (*k8sServericeClient) GetDependencies(ctx context.Context) (map[uri.URI][]*konveyor.Dep, error) {
	panic("unimplemented")
}

// GetDependenciesDAG implements provider.ServiceClient.
func (*k8sServericeClient) GetDependenciesDAG(ctx context.Context) (map[uri.URI][]konveyor.DepDAGItem, error) {
	panic("unimplemented")
}

// Stop implements provider.ServiceClient.
func (*k8sServericeClient) Stop() {
	panic("unimplemented")
}

var _ provider.ServiceClient = &k8sServericeClient{}
