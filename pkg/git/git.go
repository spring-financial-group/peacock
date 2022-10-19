package git

import (
	"github.com/spring-financial-group/peacock/pkg/domain"
	"net/url"
	"os/exec"
	"strings"
)

type Client struct {
}

func NewClient() domain.Git {
	return &Client{}
}

func (c *Client) GetRepoOwnerAndName() (string, string, error) {
	command, err := c.git("config", "--get", "remote.origin.url")
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

func (c *Client) GetLatestCommitSHA() (string, error) {
	return c.git("rev-parse", "HEAD")
}

func (c *Client) git(args ...string) (string, error) {
	e := exec.Command("git", args...)
	out, err := e.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
