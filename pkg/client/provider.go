package client

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/konveyor/analyzer-lsp/provider"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sProvider struct {
	ctx context.Context
}

// Capabilities implements provider.BaseClient.
func (*k8sProvider) Capabilities() []provider.Capability {
	return []provider.Capability{
		{
			Name: "k8s-resource-path",
		},
	}
}

// Init implements provider.BaseClient.
func (*k8sProvider) Init(ctx context.Context, log logr.Logger, initConfig provider.InitConfig) (provider.ServiceClient, error) {
	// get the kube client for the given kubeconfig, and create the service client to talk to that cluster via kubeconfig
	config, err := clientcmd.LoadFromFile(initConfig.Location)
	if err != nil {
		panic(err)
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*config, nil)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	return &k8sServericeClient{
		config: initConfig,
		client: dynamicClient,
	}, nil
}

var _ provider.BaseClient = &k8sProvider{}

func NewK8SProvider() *k8sProvider {
	return &k8sProvider{}
}
