package update

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/go-github/v20/github"
	"golang.org/x/crypto/ssh/terminal"
)

// Version exports the current version of the program
var Version = "dev"

// DefaultGitHub is the default COnfirmUpdater using the kube-switch repository as update provider
var DefaultGitHub = NewGitHub("tjamet", "kubectl-switch")

// ConfirmUpdater is the interface an object must implement to update to a the program to the latest version
// after prompting the user when ran in an interactive session
// When ran outside of a terminal, nothing is done
type ConfirmUpdater interface {
	ConfirmAndUpdate(string)
}

// Updater defines how to download a release and store it in destination
type Updater interface {
	Update(destination, source string) error
}

// Prompter defines the interface to implement to ask a command line question and read the first line
type Prompter interface {
	Prompt(string) string
}

type TTYPrompter struct {
	out io.Writer
	in  io.Reader
}

func (t TTYPrompter) Prompt(question string) string {
	_, err := t.out.Write([]byte(question))
	if err != nil {
		return ""
	}
	v, err := bufio.NewReader(t.in).ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.Trim(v, "\n\r")
}

type latestReleaseGetter interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

type NotAVersionError struct {
	Version string
	Message string
}

func (n NotAVersionError) Error() string {
	return fmt.Sprintf("%s is not a version: %s", n.Version, n.Message)
}

// Cache allows caching the update required results
// It allows reducing the impact onthe release API and get a faster response
type Cache interface {
	Set(string, bool)
	Get(string) bool
}

// HTTPUpdater implements the Updater interface to download the asset through http
type HTTPUpdater struct {
}

func (d HTTPUpdater) Update(destination, source string) error {
	r, err := http.Get(source)
	if err != nil {
		// do nothing (log?)
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// do nothing (log?)
		return err
	}
	newPath := destination + "-new"
	fd, err := os.Create(newPath)
	if err != nil {
		// do nothing (log?)
		return err
	}
	_, err = io.Copy(fd, r.Body)
	if err != nil {
		fd.Close()
		// do nothing (log?)
		return err
	}
	fd.Close()
	stat, err := os.Stat(newPath)
	if err != nil {
		// do nothing (log?)
		return err
	}
	err = os.Chmod(newPath, stat.Mode()|0111)
	if err != nil {
		// do nothing (log?)
		return err
	}
	c := exec.Command(newPath, "--help")
	err = c.Run()
	if err != nil {
		// do nothing (log?)
		return err
	}
	err = os.Rename(newPath, destination)
	if err != nil {
		// do nothing (log?)
		return err
	}
	return nil
}

// GitHub implements the ConfirmUpdater interface to update the software
// from a github repository
type GitHub struct {
	version       string
	releaseGetter latestReleaseGetter
	prompt        Prompter
	updater       Updater
	cache         Cache
	org           string
	repo          string
	os            string
	arch          string
	tty           bool
}

// NewGitHub instanciates a new GitHub ConfirmUpdater for github.com/<org>/<repo>
func NewGitHub(org, repo string) GitHub {
	return GitHub{
		version:       Version,
		releaseGetter: github.NewClient(nil).Repositories,
		prompt: TTYPrompter{
			out: os.Stdout,
			in:  os.Stdin,
		},
		updater: HTTPUpdater{},
		org:     org,
		repo:    repo,
		os:      runtime.GOOS,
		arch:    runtime.GOARCH,
		tty:     terminal.IsTerminal(int(os.Stdout.Fd())),
	}
}

func parseVersion(version string) ([]uint64, error) {
	v := strings.TrimPrefix(strings.Split(version, "-")[0], "v")
	split := []uint64{}
	for _, i := range strings.Split(v, ".") {
		u, err := strconv.ParseUint(strings.Trim(i, " "), 10, 8)
		if err != nil {
			return nil, NotAVersionError{Version: version, Message: fmt.Sprintf("%s is not a parsable: %s", i, err.Error())}
		}
		split = append(split, u)
	}
	return split, nil
}

func isNewerVersion(old, new []uint64) bool {
	for i, e := range new {
		if i > len(old)-1 {
			if e > 0 {
				return true
			}
		} else {
			if e > old[i] {
				return true
			} else if e < old[i] && i == len(old)-1 {
				return false
			}
		}
	}
	return false
}

// ConfirmAndUpdate checks the availability of a new version on github repository
// prompts the user to install it and in case the user accepts, installs it
func (g GitHub) ConfirmAndUpdate(path string) {
	release, _, err := g.releaseGetter.GetLatestRelease(context.Background(), g.org, g.repo)
	if err != nil {
		return
	}
	old, err := parseVersion(g.version)
	if err != nil {
		return
	}
	new, err := parseVersion(release.GetTagName())
	if err != nil {
		return
	}
	if isNewerVersion(old, new) {
		if g.tty {
			switch strings.ToLower(g.prompt.Prompt(fmt.Sprintf("A new version %s is available, would you like to download it? (y/N)", release.GetTagName()))) {
			case "yes", "y":
				for _, asset := range release.Assets {
					if strings.HasSuffix(asset.GetName(), fmt.Sprintf("-%s-%s", g.os, g.arch)) {
						g.updater.Update(path, asset.GetBrowserDownloadURL())
					}
				}
			}
		}
	}
}
