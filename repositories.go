package github

import (
	"context"

	gh3 "github.com/google/go-github/v32/github"
)

type Repository struct {
	Name             string   `json:"name,omitempty"`
	Private          bool     `json:"private,omitempty"`
	IsArchived       bool     `json:"is_archived,omitempty"`
	IsDisabled       bool     `json:"is_disabled,omitempty"`
	Topics           []string `json:"topics,omitempty"`
	Description      string   `json:"description,omitempty"`
	Homepage         string   `json:"homepage,omitempty"`
	Owner            string   `json:"owner,omitempty"`
	URL              string   `json:"url,omitempty"`
	Ssh              string   `json:"ssh,omitempty"`
	Git              string   `json:"git,omitempty"`
	HasIssues        bool     `json:"has_issues,omitempty"`
	HasWiki          bool     `json:"has_wiki,omitempty"`
	HasProject       bool     `json:"has_project,omitempty"`
	HasPages         bool     `json:"has_pages,omitempty"`
	IsTemplate       bool     `json:"is_template,omitempty"`
	AllowRebaseMerge bool     `json:"allow_rebase_merge,omitempty"`
	AllowSquashMerge bool     `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit bool     `json:"allow_merge_commit,omitempty"`
}

func fromRepo(r *gh3.Repository) Repository {
	return Repository{
		Name:             unboxString(r.Name),
		Private:          *r.Private,
		IsArchived:       unboxBool(r.Archived),
		IsDisabled:       unboxBool(r.Disabled),
		Topics:           r.Topics,
		Description:      unboxString(r.Description),
		Homepage:         unboxString(r.Homepage),
		Owner:            unboxString(r.Owner.Login),
		URL:              unboxString(r.URL),
		Ssh:              unboxString(r.SSHURL),
		Git:              unboxString(r.GitURL),
		HasIssues:        unboxBool(r.HasIssues),
		HasWiki:          unboxBool(r.HasWiki),
		HasProject:       unboxBool(r.HasProjects),
		HasPages:         unboxBool(r.HasPages),
		AllowRebaseMerge: unboxBool(r.AllowRebaseMerge),
		AllowSquashMerge: unboxBool(r.AllowSquashMerge),
		AllowMergeCommit: unboxBool(r.AllowMergeCommit),
	}
}

func (s *GithubClient) ListOrgRepositories(org string) ([]Repository, error) {
	var result []Repository
	err := s.paginateGithub3(func(lo gh3.ListOptions) (*gh3.Response, error) {
		repos, resp, err := s.v3Client.Repositories.ListByOrg(context.Background(), org, &gh3.RepositoryListByOrgOptions{
			ListOptions: lo,
		})
		for _, r := range repos {
			result = append(result, fromRepo(r))
		}
		return resp, err
	})
	return result, err
}

func (s *GithubClient) CreateRepository(org string, r Repository) error {
	_, _, err := s.v3Client.Repositories.Create(context.Background(), org, &gh3.Repository{
		Name:        &r.Name,
		Private:     &r.Private,
		Description: boxString(r.Description),
		Homepage:    boxString(r.Homepage),
	})
	return err
}
