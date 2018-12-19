package update

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v20/github"
	"github.com/stretchr/testify/assert"
)

type testReleaseGetter struct {
	release *github.RepositoryRelease
	err     error
	t       testing.TB
	owner   string
	repo    string
}

func (r testReleaseGetter) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	assert.Equal(r.t, r.owner, owner)
	assert.Equal(r.t, r.repo, repo)
	return r.release, nil, r.err
}

type testUpdater struct {
	source      string
	destination string
	err         error
}

func (u *testUpdater) Update(destination, source string) error {
	u.source = source
	u.destination = destination
	return u.err
}

type testPrompter struct {
	answer string
}

func (t testPrompter) Prompt(question string) string {
	return t.answer
}

type testFailerPrompter struct {
	t testing.TB
}

func (t testFailerPrompter) Prompt(question string) string {
	t.t.Errorf("The prompter should not be called")
	return ""
}

func testConfirmAndUpdate(t testing.TB, oldVersion, expectedDest, expectedSource string, releaseGetter latestReleaseGetter, prompt Prompter, tty bool) {
	updater := &testUpdater{}
	g := GitHub{
		updater:       updater,
		os:            "windows",
		arch:          "world",
		prompt:        prompt,
		releaseGetter: releaseGetter,
		tty:           tty,
		version:       oldVersion,
	}
	g.ConfirmAndUpdate("hello-world")
	assert.Equal(t, expectedSource, updater.source)
	assert.Equal(t, expectedDest, updater.destination)
}

func testConfirmAndUpdateWithRelease(t testing.TB, oldVersion, newVersion, expectedDest, expectedSource string, prompt Prompter, tty bool) {
	defer func(v string) {
		Version = v
	}(Version)
	Version = oldVersion
	release := &github.RepositoryRelease{
		TagName: github.String(newVersion),
		Assets: []github.ReleaseAsset{
			github.ReleaseAsset{
				BrowserDownloadURL: github.String("file-<os>-<arch>"),
				Name:               github.String("file-<os>-<arch>"),
			},
			github.ReleaseAsset{
				BrowserDownloadURL: github.String("http://localhost/<hash>"),
				Name:               github.String("file-windows-world"),
			},
		},
	}
	testConfirmAndUpdate(t, oldVersion, expectedDest, expectedSource, testReleaseGetter{release: release}, prompt, tty)
}

func TestGitHubGetVersion(t *testing.T) {
	t.Run("when getting a release fails, the program is not updated", func(t *testing.T) {
		testConfirmAndUpdate(t, "", "", "", testReleaseGetter{err: fmt.Errorf("test error")}, testPrompter{}, true)
	})
	t.Run("when the old release is unparsable, the program is not updated", func(t *testing.T) {
		testConfirmAndUpdateWithRelease(t, "", "", "", "", testPrompter{}, true)
	})
	t.Run("when the new release is unparsable, the program is not updated", func(t *testing.T) {
		testConfirmAndUpdateWithRelease(t, "v1.5", "", "", "", testPrompter{}, true)
	})
	t.Run("when the new release is the same as the latest version, the program is not updated", func(t *testing.T) {
		testConfirmAndUpdateWithRelease(t, "v1.5", "1.5", "", "", testPrompter{}, true)
	})
	for _, answer := range []string{"", "n", "N", "no", "No", "something"} {
		t.Run("when the user refuses the update, the program is not updated", func(t *testing.T) {
			testConfirmAndUpdateWithRelease(t, "v1.5", "1.5.3", "", "", testPrompter{answer: answer}, true)
		})
	}
	for _, answer := range []string{"y", "yes", "YES", "Y"} {
		t.Run("when the user refuses the update, the program is not updated", func(t *testing.T) {
			testConfirmAndUpdateWithRelease(t, "v1.5", "1.5.3", "hello-world", "http://localhost/<hash>", testPrompter{answer: answer}, true)
		})
	}

	t.Run("when the updater is not running on a TTY", func(t *testing.T) {
		testConfirmAndUpdateWithRelease(t, "v1.5", "1.5.3", "", "", testFailerPrompter{t: t}, false)
	})
}

func TestNotAVersionError(t *testing.T) {
	assert.Equal(t, "<version> is not a version: <message>", NotAVersionError{
		Version: "<version>",
		Message: "<message>",
	}.Error())
}

func TestIsNewerVersion(t *testing.T) {
	assert.True(t, isNewerVersion([]uint64{}, []uint64{1}))
	assert.True(t, isNewerVersion([]uint64{1}, []uint64{1, 1}))
	assert.True(t, isNewerVersion([]uint64{1, 9}, []uint64{1, 10}))
	assert.False(t, isNewerVersion([]uint64{1}, []uint64{1}))
	assert.False(t, isNewerVersion([]uint64{2}, []uint64{1}))
	assert.False(t, isNewerVersion([]uint64{1, 4}, []uint64{1}))
	assert.False(t, isNewerVersion([]uint64{2}, []uint64{1, 9}))
}

func TestVersion(t *testing.T) {
	testVersion(t, []uint64{1, 2, 3}, "v1.2.3")
	testVersion(t, []uint64{1, 2, 3, 4}, "v1.2.3.4-dev")
	testVersionInError(t, "")
	testVersionInError(t, "dev")
}

func testVersion(t testing.TB, expected []uint64, version string) {
	v, err := parseVersion(version)
	assert.NoError(t, err)
	assert.Equal(t, expected, v)
}

func testVersionInError(t testing.TB, version string) {
	v, err := parseVersion(version)
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestHTTPUpdater(t *testing.T) {
	defer os.Remove("test-updated")
	defer os.Remove("test-not-updated")
	defer os.Remove("test-updated-new")
	defer os.Remove("test-not-updated-new")
	t.Run("when the help message returned by the server succeeds, installation succeeds", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/asset", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("#!/bin/bash\necho $@\n[ \"x$@\" == \"x--help\" ]"))
		}))
		defer s.Close()
		HTTPUpdater{}.Update("./test-updated", s.URL+"/asset")
		assert.FileExists(t, "./test-updated")
	})
	t.Run("when the server fails, installation fails", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/asset", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("#!/bin/bash\nexit 1"))
		}))
		defer s.Close()
		HTTPUpdater{}.Update("./test-not-updated", s.URL+"/asset")
		_, err := os.Stat("./test-not-updated")
		assert.True(t, os.IsNotExist(err))
	})
	t.Run("when the server fails, old installation remains", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/asset", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("#!/bin/bash\nexit 1"))
		}))
		defer s.Close()
		_, err := os.Create("./test-not-updated")
		assert.NoError(t, err)
		HTTPUpdater{}.Update("./test-not-updated", s.URL+"/asset")
		assert.FileExists(t, "./test-not-updated")
	})
}

type testReaderError struct{}

func (t testReaderError) Read([]byte) (int, error) {
	return 0, fmt.Errorf("test error")
}

type testWriter struct {
	t   testing.TB
	d   string
	n   int
	err error
}

func (t testWriter) Write(d []byte) (int, error) {
	assert.Equal(t.t, t.d, string(d))
	return t.n, t.err
}

func TestTTYPrompter(t *testing.T) {
	assert.Equal(
		t,
		"some word",
		TTYPrompter{
			in: strings.NewReader("some word\nignored line"),
			out: testWriter{
				t: t,
				d: "hello world?",
			},
		}.Prompt("hello world?"))
	assert.Equal(
		t,
		"",
		TTYPrompter{
			out: testWriter{
				t:   t,
				d:   "hello world?",
				err: fmt.Errorf("test error"),
			},
		}.Prompt("hello world?"))
	assert.Equal(
		t,
		"",
		TTYPrompter{
			in:  testReaderError{},
			out: bytes.NewBuffer(nil),
		}.Prompt("hello world?"))
}

//integration test to check github integration
func TestDefaultGitHub(t *testing.T) {
	gh := DefaultGitHub
	updater := &testUpdater{}
	gh.prompt = testPrompter{answer: "y"}
	gh.updater = updater
	gh.version = "0"
	gh.tty = true
	gh.ConfirmAndUpdate("./kubectl")
	assert.NotEqual(t, "", updater.source)
	assert.Equal(t, "./kubectl", updater.destination)
}
