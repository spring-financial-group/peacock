package git

import (
	"net/url"
	"os/exec"
	"strings"
)

type Client struct {
}

func NewClient() Client {
	return Client{}
}

func (c Client) GetRepoOwnerAndName(dir string) (string, string, error) {
	command, err := c.git(dir, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", "", err
	}
	url, err := url.Parse(command)
	if err != nil {
		return "", "", err
	}
	path := strings.TrimSuffix(url.Path, ".git")
	path = strings.TrimPrefix(path, "/")
	split := strings.Split(path, "/")
	return split[0], split[1], nil
}

func (c Client) GetLatestCommitSHA(dir string) (string, error) {
	return c.git(dir, "rev-parse", "HEAD")
}

func (c Client) git(dir string, args ...string) (string, error) {
	e := exec.Command("git", args...)
	e.Dir = dir
	out, err := e.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
