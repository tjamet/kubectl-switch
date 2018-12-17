package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
)

type testRestConfigGetterFunc func() (*rest.Config, error)

func (t testRestConfigGetterFunc) ToRESTConfig() (*rest.Config, error) {
	return t()
}

func setNewKubernetesClient(f func(c *rest.Config) (discovery.ServerVersionInterface, error)) {
	newKubernetesClient = f
}

func setDefaultVersion(v string){
	DefaultVersion = v
}

func setTimeout(d time.Duration){
	Timeout = d
}

type testServerVersion struct {
	info *version.Info
	err  error
}
func (t testServerVersion) ServerVersion() (*version.Info, error) {
	return t.info, t.err
}

func TestGetVersion(t *testing.T) {
	defer setNewKubernetesClient(newKubernetesClient)
	defer setDefaultVersion(DefaultVersion)
	defer setTimeout(Timeout)
	setTimeout(3 * time.Second)

	t.Run("when version is correctly returned, this default version is returned", func(t *testing.T){
		config := &rest.Config{
			Host: "http://test",
		}
		setNewKubernetesClient(func(c*rest.Config) (discovery.ServerVersionInterface, error){
			assert.Equal(t, config, c)
			return testServerVersion{info:&version.Info{GitVersion:"1.10-test"}}, nil
		})
		assert.Equal(t, "1.10-test", GetVersionFromConfig(testRestConfigGetterFunc(func() (*rest.Config, error) {
			return config, nil
		})))
	})


	setDefaultVersion("test-default-version")

	t.Run("when config getter succeeds, this default version is returned", func(t *testing.T){
		setNewKubernetesClient(func(*rest.Config) (discovery.ServerVersionInterface, error){
			return testServerVersion{err:fmt.Errorf("test error")}, nil
		})
		assert.Equal(t, "test-default-version", GetVersionFromConfig(testRestConfigGetterFunc(func() (*rest.Config, error) {
			return nil, nil
		})))
	})

	t.Run("when client getter is in error, the default version is returned", func(t *testing.T){
		setNewKubernetesClient(func(*rest.Config) (discovery.ServerVersionInterface, error){
			return nil, fmt.Errorf("this is a test error")
		})
		assert.Equal(t, "test-default-version", GetVersionFromConfig(testRestConfigGetterFunc(func() (*rest.Config, error) {
			return nil, nil
		})))
	})

	t.Run("when config getter fails, the default version is returned", func(t *testing.T){
		assert.Equal(t, "test-default-version", GetVersionFromConfig(testRestConfigGetterFunc(func() (*rest.Config, error) {
			return nil, fmt.Errorf("new test error")
		})))
	})
	t.Run("when config getter is taking too long, the default version is returned", func(t *testing.T){
		setTimeout(1 * time.Nanosecond)
		assert.Equal(t, "test-default-version", GetVersionFromConfig(testRestConfigGetterFunc(func() (*rest.Config, error) {
			<- time.After(30 * time.Second)
			return nil, nil
		})))
	})

}
