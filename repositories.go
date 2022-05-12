package github

import (
	"context"
	"time"

	log "github.com/mtrense/soil/logging"

	"github.com/shurcooL/githubv4"

	gh3 "github.com/google/go-github/v32/github"
)

type Repository struct {
	Name                  string                 `json:"name,omitempty"`
	Private               bool                   `json:"private,omitempty"`
	Description           string                 `json:"description,omitempty"`
	Homepage              string                 `json:"homepage,omitempty"`
	Owner                 string                 `json:"owner,omitempty"`
	URL                   string                 `json:"url,omitempty"`
	Ssh                   string                 `json:"ssh,omitempty"`
	Git                   string                 `json:"git,omitempty"`
	DefaultBranch         string                 `json:"default_branch,omitempty"`
	Archived              bool                   `json:"archived,omitempty"`
	Disabled              bool                   `json:"disabled,omitempty"`
	Template              bool                   `json:"template,omitempty"`
	IssuesEnabled         bool                   `json:"issues_enabled,omitempty"`
	WikiEnabled           bool                   `json:"wiki_enabled,omitempty"`
	ProjectsEnabled       bool                   `json:"projects_enabled,omitempty"`
	PagesEnabled          bool                   `json:"pages_enabled,omitempty"`
	AllowRebaseMerge      bool                   `json:"allow_rebase_merge,omitempty"`
	AllowSquashMerge      bool                   `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit      bool                   `json:"allow_merge_commit,omitempty"`
	DeleteBranchOnMerge   bool                   `json:"delete_branch_on_merge,omitempty"`
	VulnerabilityAlerts   bool                   `json:"vulnerability_alerts,omitempty"`
	LicenseName           string                 `json:"license_name,omitempty"`
	Topics                []string               `json:"topics,omitempty"`
	Collaborators         []Collaborator         `json:"collaborators,omitempty"`
	BranchProtectionRules []BranchProtectionRule `json:"branch_protection_rules,omitempty"`
	DiskUsage             int                    `json:"disk_usage,omitempty"`
	PushedAt              time.Time              `json:"pushed_at,omitempty"`
	UpdatedAt             time.Time              `json:"updated_at,omitempty"`
	Age                   int                    `json:"age,omitempty"`
	PrimaryLanguage       string                 `json:"primary_language,omitempty"`
	Languages             []Language             `json:"languages,omitempty"`
	Workflows             []Workflow             `json:"workflows,omitempty"`
}

type Language struct {
	Name        string `json:"name,omitempty"`
	LinesOfCode int    `json:"lines_of_code,omitempty"`
}

func (s *GithubClient) LoadRepositoryLanguages(repositories ...*Repository) error {
	for _, repository := range repositories {
		if err := s.loadRepositoryLanguages(repository); err != nil {
			return err
		}
	}
	return nil
}

func (s *GithubClient) loadRepositoryLanguages(repository *Repository) error {
	var query struct {
		Repository struct {
			Languages struct {
				PageInfo PageInfo
				Edges    []struct {
					Node struct {
						Name githubv4.String
					}
					Size githubv4.Int
				}
			} `graphql:"languages(first: 100, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	return s.Query(&query).Str("owner", repository.Owner).Str("repo", repository.Name).Cursor("cursor").RunPaginated(func() PageInfo {
		for _, lang := range query.Repository.Languages.Edges {
			language := Language{
				Name:        string(lang.Node.Name),
				LinesOfCode: int(lang.Size),
			}
			repository.Languages = append(repository.Languages, language)
		}
		return query.Repository.Languages.PageInfo
	})
}

type Collaborator struct {
	Login               string             `json:"login,omitempty"`
	EffectivePermission string             `json:"effective_permission,omitempty"`
	Sources             []PermissionSource `json:"sources,omitempty"`
}

type PermissionSource struct {
	Organization    string `json:"organization,omitempty"`
	Permission      string `json:"permission,omitempty"`
	SourceType      string `json:"source_type,omitempty"`
	TeamName        string `json:"team_name,omitempty"`
	TeamID          string `json:"team_id,omitempty"`
	RepositoryOwner string `json:"repository_owner,omitempty"`
	RepositoryName  string `json:"repository_name,omitempty"`
}

func (s *GithubClient) LoadRepositoryCollaborators(repositories ...*Repository) error {
	for _, repository := range repositories {
		if !repository.Archived {
			if err := s.loadRepositoryCollaborators(repository); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *GithubClient) loadRepositoryCollaborators(repository *Repository) error {
	var query struct {
		Repository struct {
			Collaborators struct {
				PageInfo PageInfo
				Edges    []struct {
					Node struct {
						Login githubv4.String
					}
					Permission        githubv4.String
					PermissionSources []struct {
						Organization struct {
							Name githubv4.String
						}
						Permission githubv4.String
						Source     struct {
							Typename githubv4.String `graphql:"__typename"`
							Team     struct {
								Name githubv4.String
								ID   githubv4.String
							} `graphql:"... on Team"`
							Repository struct {
								Owner struct {
									Login githubv4.String
								}
								Name githubv4.String
							} `graphql:"... on Repository"`
							Organization struct {
								Name githubv4.String
							} `graphql:"... on Organization"`
						}
					}
				}
			} `graphql:"collaborators(first: 10, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	log.L().Info().Str("repo", repository.Name).Msg("Fetching Collaborators")
	return s.Query(&query).Str("owner", repository.Owner).Str("repo", repository.Name).Cursor("cursor").RunPaginated(func() PageInfo {
		log.L().Info().Bool("next-page", bool(query.Repository.Collaborators.PageInfo.HasNextPage)).Str("repo", repository.Name).Msg("Fetched next page")
		for _, coll := range query.Repository.Collaborators.Edges {
			var sources []PermissionSource
			for _, ps := range coll.PermissionSources {
				permissionSource := PermissionSource{
					Organization:    string(ps.Organization.Name),
					Permission:      string(ps.Permission),
					SourceType:      string(ps.Source.Typename),
					TeamName:        string(ps.Source.Team.Name),
					TeamID:          string(ps.Source.Team.ID),
					RepositoryOwner: string(ps.Source.Repository.Owner.Login),
					RepositoryName:  string(ps.Source.Repository.Name),
				}
				sources = append(sources, permissionSource)
			}
			collaborator := Collaborator{
				Login:               string(coll.Node.Login),
				EffectivePermission: string(coll.Permission),
				Sources:             sources,
			}
			repository.Collaborators = append(repository.Collaborators, collaborator)
		}
		return query.Repository.Collaborators.PageInfo
	})
}

type BranchProtectionRule struct {
	Pattern                      string   `json:"pattern,omitempty"`
	MatchingRefs                 []string `json:"matching_refs,omitempty"`
	AllowsForcePushes            bool     `json:"allows_force_pushes,omitempty"`
	AllowsDeletions              bool     `json:"allows_deletions,omitempty"`
	RequiredApprovingReviewCount int      `json:"required_approving_review_count,omitempty"`
	RequiredStatusCheckContexts  []string `json:"required_status_check_contexts,omitempty"`
	RequiresApprovingReviews     bool     `json:"requires_approving_reviews,omitempty"`
	RequiresCodeOwnerReviews     bool     `json:"requires_code_owner_reviews,omitempty"`
	RequiresCommitSignatures     bool     `json:"requires_commit_signatures,omitempty"`
	RequiresLinearHistory        bool     `json:"requires_linear_history,omitempty"`
	RequiresStrictStatusChecks   bool     `json:"requires_strict_status_checks,omitempty"`
	IsAdminEnforced              bool     `json:"is_admin_enforced,omitempty"`
	RestrictsReviewDismissals    bool     `json:"restricts_review_dismissals,omitempty"`
	DismissesStaleReviews        bool     `json:"dismisses_stale_reviews,omitempty"`
}

func (s *GithubClient) LoadRepositoryBranchProtectionRules(repositories ...*Repository) error {
	for _, repository := range repositories {
		if err := s.loadRepositoryBranchProtectionRules(repository); err != nil {
			return err
		}
	}
	return nil
}

func (s *GithubClient) loadRepositoryBranchProtectionRules(repository *Repository) error {
	var query struct {
		Repository struct {
			BranchProtectionRules struct {
				PageInfo PageInfo
				Nodes    []struct {
					Pattern      githubv4.String
					MatchingRefs struct {
						Nodes []struct {
							Name githubv4.String
						}
					} `graphql:"matchingRefs(first: 100)"`
					AllowsForcePushes            githubv4.Boolean
					AllowsDeletions              githubv4.Boolean
					RequiredApprovingReviewCount githubv4.Int
					RequiredStatusCheckContexts  []githubv4.String
					RequiresApprovingReviews     githubv4.Boolean
					RequiresCodeOwnerReviews     githubv4.Boolean
					RequiresCommitSignatures     githubv4.Boolean
					RequiresLinearHistory        githubv4.Boolean
					RequiresStrictStatusChecks   githubv4.Boolean
					IsAdminEnforced              githubv4.Boolean
					RestrictsReviewDismissals    githubv4.Boolean
					DismissesStaleReviews        githubv4.Boolean
				}
			} `graphql:"branchProtectionRules(first: 100, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	return s.Query(&query).Str("owner", repository.Owner).Str("repo", repository.Name).Cursor("cursor").RunPaginated(func() PageInfo {
		log.L().Info().Bool("next-page", bool(query.Repository.BranchProtectionRules.PageInfo.HasNextPage)).Str("repo", repository.Name).Msg("Fetched next page")
		for _, r := range query.Repository.BranchProtectionRules.Nodes {
			var matchingRefs []string
			for _, ref := range r.MatchingRefs.Nodes {
				matchingRefs = append(matchingRefs, string(ref.Name))
			}
			var statusChecks []string
			for _, ref := range r.RequiredStatusCheckContexts {
				statusChecks = append(statusChecks, string(ref))
			}
			rule := BranchProtectionRule{
				Pattern:                      string(r.Pattern),
				MatchingRefs:                 matchingRefs,
				AllowsForcePushes:            bool(r.AllowsForcePushes),
				AllowsDeletions:              bool(r.AllowsDeletions),
				RequiredApprovingReviewCount: int(r.RequiredApprovingReviewCount),
				RequiredStatusCheckContexts:  statusChecks,
				RequiresApprovingReviews:     bool(r.RequiresApprovingReviews),
				RequiresCodeOwnerReviews:     bool(r.RequiresCodeOwnerReviews),
				RequiresCommitSignatures:     bool(r.RequiresCommitSignatures),
				RequiresLinearHistory:        bool(r.RequiresLinearHistory),
				RequiresStrictStatusChecks:   bool(r.RequiresStrictStatusChecks),
				IsAdminEnforced:              bool(r.IsAdminEnforced),
				RestrictsReviewDismissals:    bool(r.RestrictsReviewDismissals),
				DismissesStaleReviews:        bool(r.DismissesStaleReviews),
			}
			repository.BranchProtectionRules = append(repository.BranchProtectionRules, rule)
		}
		return query.Repository.BranchProtectionRules.PageInfo
	})
}

func (s *GithubClient) LoadRepositorySecurityConfig(repositories ...*Repository) error {
	for _, repository := range repositories {
		if err := s.loadSecurityConfig(repository); err != nil {
			return err
		}
	}
	return nil
}

func (s *GithubClient) loadSecurityConfig(repository *Repository) error {
	if enabled, _, err := s.v3Client.Repositories.GetVulnerabilityAlerts(context.Background(), repository.Owner, repository.Name); err == nil {
		repository.VulnerabilityAlerts = enabled
	} else {
		return err
	}
	return nil
}

func (s *GithubClient) EnableVulnerabilityAlerts(owner string, repository string) error {
	_, err := s.v3Client.Repositories.EnableVulnerabilityAlerts(context.Background(), owner, repository)
	return err
}

type Workflow struct {
	Name         string `json:"name,omitempty"`
	Path         string `json:"path,omitempty"`
	State        string `json:"state,omitempty"`
	UsageWindows int64  `json:"usage_windows,omitempty"`
	UsageUbuntu  int64  `json:"usage_ubuntu,omitempty"`
	UsageMac     int64  `json:"usage_mac,omitempty"`
}

func (s *GithubClient) LoadRepositoryWorkflows(repositories ...*Repository) error {
	for _, repository := range repositories {
		if err := s.loadRepositoryWorkflows(repository); err != nil {
			return err
		}
	}
	return nil
}

func (s *GithubClient) loadRepositoryWorkflows(repository *Repository) error {
	return s.paginateGithub3(func(lo gh3.ListOptions) (*gh3.Response, error) {
		workflows, resp, err := s.v3Client.Actions.ListWorkflows(context.Background(), repository.Owner, repository.Name, &lo)
		for _, w := range workflows.Workflows {
			usage, _, err := s.v3Client.Actions.GetWorkflowUsageByID(context.Background(), repository.Owner, repository.Name, w.GetID())
			if err != nil {
				return nil, err
			}
			repository.Workflows = append(repository.Workflows, Workflow{
				Name:         w.GetName(),
				Path:         w.GetPath(),
				State:        w.GetState(),
				UsageMac:     usage.GetBillable().GetMacOS().GetTotalMS(),
				UsageUbuntu:  usage.GetBillable().GetUbuntu().GetTotalMS(),
				UsageWindows: usage.GetBillable().GetWindows().GetTotalMS(),
			})
		}
		return resp, err
	})
}

func (s *GithubClient) CreateRepository(org string, r *Repository) error {
	_, _, err := s.v3Client.Repositories.Create(context.Background(), org, &gh3.Repository{
		Name:        &r.Name,
		Private:     &r.Private,
		Description: boxString(r.Description),
		Homepage:    boxString(r.Homepage),
	})
	return err
}
