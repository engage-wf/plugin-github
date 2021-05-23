package github

import (
	"context"

	gh3 "github.com/google/go-github/v32/github"
)

type PublicKey struct {
	ID    *int64  `json:"id,omitempty"`
	Key   *string `json:"key,omitempty"`
	URL   *string `json:"url,omitempty"`
	Title *string `json:"title,omitempty"`
}

func fromGh3Key(key *gh3.Key) PublicKey {
	return PublicKey{
		ID:    key.ID,
		Key:   key.Key,
		URL:   key.URL,
		Title: key.Title,
	}
}

func (s *GithubClient) ListPublicKeys(user string) ([]PublicKey, error) {
	var result []PublicKey
	err := s.paginateGithub3(func(lo gh3.ListOptions) (*gh3.Response, error) {
		keys, resp, err := s.v3Client.Users.ListKeys(context.Background(), user, &lo)
		for _, key := range keys {
			result = append(result, fromGh3Key(key))
		}
		return resp, err
	})
	return result, err
}
