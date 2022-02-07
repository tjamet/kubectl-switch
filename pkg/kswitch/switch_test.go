package kswitch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing" //"time"

	"github.com/stretchr/testify/assert"
	"github.com/tjamet/kubectl-switch/pkg/kubectl"
	//"github.com/tjamet/kubectl-switch/server"
)

const expectedOutput = "hello world"
const envVarName = "K_PROXY_TEST_MAIN"

var envVarMagic = "this is magic"

func testMain(t *testing.T, d string, args []string) {
	t.Helper()
	os.Args = args
	os.Setenv(envVarName, envVarMagic)
	stdout, err := os.Create(filepath.Join(d, "stdout"))
	assert.NoError(t, err)
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(d, "stderr"))
	assert.NoError(t, err)
	defer stderr.Close()
	defer func(o, e *os.File) {
		os.Stdout = o
		os.Stderr = e
	}(os.Stdout, os.Stderr)
	os.Stdout = stdout
	os.Stderr = stderr
	exit = func(code int) {
		assert.Equal(t, 0, code)
	}
	Main()
	fd, err := os.Open(filepath.Join(d, "stdout"))
	assert.NoError(t, err)
	defer fd.Close()
	b, err := ioutil.ReadAll(fd)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput+"\n", string(b))
	fd, err = os.Open(filepath.Join(d, "stderr"))
	assert.NoError(t, err)
	defer fd.Close()
	b, err = ioutil.ReadAll(fd)
	assert.NoError(t, err)
	assert.Equal(t, "", string(b))
}

func TestIntegration(t *testing.T) {
	t.Cleanup(kubectl.Reset)
	d, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(d)
	})
	kubectl.HomeDir = func() string { return d }
	path := kubectl.Path("1.13.1")
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0777))
	assert.NoError(t, os.Link(os.Args[0], path))
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"major": "1",
			"minor": "13",
			"gitVersion": "v1.13.1",
			"gitCommit": "eec55b9ba98609a46fee712359c7b5b365bdd920",
			"gitTreeState": "clean",
			"buildDate": "2018-12-13T10:31:33Z",
			"goVersion": "go1.11.2",
			"compiler": "gc",
			"platform": "linux/amd64"
		  }`))
	}))
	testMain(t, d, []string{"test", "--server", s.URL, "--token", "test-token", "--unknown", "-o", "something"})
	testMain(t, d, []string{"test", "--server", s.URL, "--token", "test-token", "--help"})
}

func TestMain(m *testing.M) {
	if os.Getenv(envVarName) == envVarMagic {
		fmt.Println(expectedOutput)
		os.Exit(0)
	}
	os.Exit(m.Run())
}
