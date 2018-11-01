package kubectl_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tjamet/kubectl-switch/kubectl"
	"github.com/tjamet/xgo/xtesting"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}

func testURL(t testing.TB, version, expectedURL string) {
	assert.Equal(t, expectedURL, kubectl.URL(version))
}

func mockKubectl(t testing.TB, code int) func() {
	defaultTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "storage.googleapis.com", r.Host)
		assert.Contains(t, "kubectl", r.RequestURI)
		return &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf("#!/bin/bash\necho $@\nexit %d\n", code))),
			StatusCode: http.StatusOK,
		}, nil
	})
	return func() {
		http.DefaultTransport = defaultTransport
	}
}

func TestURL(t *testing.T) {
	testURL(t, "v1.0.0", fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v1.0.0/bin/%s/%s/kubectl", runtime.GOOS, runtime.GOARCH))
	testURL(t, "1.0.0", fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v1.0.0/bin/%s/%s/kubectl", runtime.GOOS, runtime.GOARCH))
	testURL(t, "v1.0.0+coreos", fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v1.0.0/bin/%s/%s/kubectl", runtime.GOOS, runtime.GOARCH))
}

func TestPath(t *testing.T) {
	defer xtesting.NoEnv("HOME")()
	defer xtesting.InEnv("USERPROFILE", "./test-home")()
	assert.Equal(t, "./test-home/.kube/bin/kubectl-1.10.0", kubectl.Path("1.10.0"))
}

func TestDownload(t *testing.T) {
	defer xtesting.InEnv("HOME", "./test-home")()
	assert.False(t, kubectl.Installed("test-some-version"))
	t.Run("When the server returns a 404, kubectl is not installed", func(t *testing.T) {
		defaultTransport := http.DefaultTransport
		http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("#!/bin/bash\necho $@\nexit 1\n")),
				StatusCode: http.StatusNotFound,
			}, nil
		})
		defer func() {
			http.DefaultTransport = defaultTransport
		}()
		assert.Error(t, kubectl.Download("v0.0.9.9"))
		assert.False(t, kubectl.Installed("v0.0.9.9"))
	})
}

func TestDownloadAndRun(t *testing.T) {
	defer xtesting.InEnv("HOME", "./test-home")()
	assert.False(t, kubectl.Installed("test-some-version"))
	t.Run("When kubectl returns an error, Run returns the status code", func(t *testing.T) {
		defer mockKubectl(t, 1)()
		assert.NoError(t, kubectl.Download("0.0.0.1"))
		assert.FileExists(t, "./test-home/.kube/bin/kubectl-0.0.0.1")
		assert.Equal(t, 1, kubectl.Exec("0.0.0.1"))
		assert.True(t, kubectl.Installed("0.0.0.1"))
	})
	t.Run("When kubectl returns an error, Run returns the status code", func(t *testing.T) {
		defer mockKubectl(t, 0)()
		assert.NoError(t, kubectl.Download("0.0.0.2"))
		assert.FileExists(t, "./test-home/.kube/bin/kubectl-0.0.0.2")
		assert.Equal(t, 0, kubectl.Exec("0.0.0.2"))
	})
	t.Run("Arguments are forwarded to kubectl", func(t *testing.T) {
		defer mockKubectl(t, 0)()
		out, err := kubectl.Command("0.0.0.2", "version", "--help").Output()
		assert.NoError(t, err)
		assert.Contains(t, string(out), "version")
		assert.Contains(t, string(out), "--help")
	})
}
