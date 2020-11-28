package github

import (
	"context"

	gh3 "github.com/google/go-github/v32/github"
	gh4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	v4Client *gh4.Client
	v3Client *gh3.Client
}

func New(token string) *GithubClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	return &GithubClient{
		v4Client: gh4.NewClient(httpClient),
		v3Client: gh3.NewClient(httpClient),
	}
}

func (s *GithubClient) paginateGithub3(fetcher func(lo gh3.ListOptions) (*gh3.Response, error)) error {
	listOptions := gh3.ListOptions{
		Page:    0,
		PerPage: 100,
	}
	for {
		resp, err := fetcher(listOptions)
		if err != nil {
			return err
		}
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page += 1
	}
	return nil
}
