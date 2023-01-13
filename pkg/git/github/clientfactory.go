package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"golang.org/x/oauth2"
)

type ClientFactory struct {
	github  *github.Client
	clients map[string]*Client
}

func NewClientFactory(token string) *ClientFactory {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &ClientFactory{
		github:  github.NewClient(tc),
		clients: make(map[string]*Client),
	}
}

func (c *ClientFactory) GetClient(owner, repo, user string, prNumber int) domain.SCM {
	key := fmt.Sprintf("%s/%s/%s/%d", user, owner, repo, prNumber)
	if client, ok := c.clients[key]; ok {
		return client
	}
	client := &Client{
		github:   c.github,
		owner:    owner,
		repo:     repo,
		prNumber: prNumber,
	}
	c.clients[key] = client
	return client
}

func (c *ClientFactory) RemoveClient(key string) {
	delete(c.clients, key)
}
