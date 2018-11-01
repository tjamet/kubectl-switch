package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"syscall"
	"text/template"
)

// KubectlURLTemplate defines the pattern of download URLs for kubectl
// Edit this value to use a custom location
var KubectlURLTemplate = "https://storage.googleapis.com/kubernetes-release/release/v{{ .Version }}/bin/{{ .OS }}/{{ .Arch }}/kubectl"
var versionRegexp = regexp.MustCompile("([0-9.]+)")

type kubectlVersion struct {
	Version string
	OS      string
	Arch    string
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func normalizeVersion(version string) string {
	versionRegexp.Longest()
	matches := versionRegexp.FindAllString(version, 1)
	if len(matches) >= 1 {
		return matches[0]
	}
	return ""
}

// URL returns the URL where a given kubectl version should be downloaded from
func URL(version string) string {
	version = normalizeVersion(version)
	t := template.Must(template.New("URL").Parse(KubectlURLTemplate))
	b := bytes.NewBuffer(nil)
	t.Execute(b, &kubectlVersion{
		Version: version,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	})
	return b.String()
}

func binDir() string {
	return homeDir() + "/.kube/bin"
}

// Path retrieves the path of kubectl for a given version
func Path(version string) string {
	version = normalizeVersion(version)
	return binDir() + "/kubectl-" + version
}

// Download downloads a specific kubectl version to Path()
func Download(version string) error {
	version = normalizeVersion(version)

	kubectlURL := URL(version)
	kubectl := Path(version)

	if err := os.MkdirAll(binDir(), 0766); err != nil {
		return fmt.Errorf("failed to create bin directory %s: %s", binDir(), err)
	}

	fmt.Fprintf(os.Stderr, "Downloading kubectl from %s\n", kubectlURL)
	response, err := http.Get(kubectlURL)
	if err != nil {
		return fmt.Errorf("failed to download the kubectl version: %s", err)
	}
	defer response.Body.Close()
	fd, err := os.Create(kubectl)
	if err != nil {
		return fmt.Errorf("failed to write the kubectl version: %s", err)
	}
	defer fd.Close()
	_, err = io.Copy(fd, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write the kubectl version: %s", err)
	}
	if err := os.Chmod(kubectl, 0766); err != nil {
		return fmt.Errorf("failed to set execution permissions for kubectl: %s", err)
	}
	return nil
}

// Installed returns wether the kubectl version is already installed or not
func Installed(version string) bool {
	version = normalizeVersion(version)
	_, err := os.Stat(Path(version))
	return !os.IsNotExist(err)
}

// Command instanciates a new command for a given version
func Command(version string, args ...string) *exec.Cmd {
	version = normalizeVersion(version)
	return exec.Command(Path(version), args...)
}

// Run executes the Command, binding standard input and outputs
func Run(version string, args ...string) error {
	version = normalizeVersion(version)
	e := Command(version, args...)
	e.Stdin = os.Stdin
	e.Stderr = os.Stderr
	e.Stdout = os.Stdout
	return e.Run()
}

// Exec executes the Command and returns the status code in case of an error
func Exec(version string, args ...string) int {
	version = normalizeVersion(version)
	if err := Run(version, args...); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		} else {
			fmt.Println(err)
			return 1
		}
	}
	return 0
}
