package server

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/discovery"
)

// DefaultVersion is the client version to use when failing to get the server one
var DefaultVersion = "1.13.0"

// Timeout is the maximum time to retrieve the version
var Timeout = 1 * time.Second

// allow dependency injection for tests
var newKubernetesClient = func(c *rest.Config) (discovery.ServerVersionInterface, error) {
	return kubernetes.NewForConfig(c)
}

// RestConfigGetter defines the interface a rest configuration loader must implement
// to be able to get a given version
type RestConfigGetter interface {
	ToRESTConfig() (*rest.Config, error)
}

// GetVersionFromConfig retrieves the configuration file and returns the version a server implements
// If the server is unreachable or in case of an error, it defaults to the latest version
func GetVersionFromConfig(rg RestConfigGetter) string {
	v := make(chan string)
	go func(v chan string) {
		config, err := rg.ToRESTConfig()
		if err != nil {
			v <- DefaultVersion
			return
		}
		clientset, err := newKubernetesClient(config)
		if err != nil {
			v <- DefaultVersion
			return
		}
		version, err := clientset.ServerVersion()
		if err != nil {
			v <- DefaultVersion
			return
		}
		v <- version.String()
	}(v)
	select {
	case version := <-v:
		return version
	case <-time.After(Timeout):
		return DefaultVersion
	}
}
