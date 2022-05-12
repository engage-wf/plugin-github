package github

import (
	"time"

	"github.com/shurcooL/githubv4"
)

type Organization struct {
	Name string
}

type Member struct {
	Login               string    `json:"login,omitempty"`
	Name                string    `json:"name,omitempty"`
	Email               string    `json:"email,omitempty"`
	HasTwoFactorEnabled bool      `json:"has_two_factor_enabled,omitempty"`
	Role                string    `json:"role,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	Pending             bool      `json:"pending"`
}

type MemberOption func()

func (s *GithubClient) GetMembers(org string, opts ...MemberOption) ([]*Member, error) {
	var members []*Member
	if nonPendingMembers, err := s.getNonPendingMembers(org); err == nil {
		for _, m := range nonPendingMembers {
			members = append(members, m)
		}
	} else {
		return nil, err
	}
	if pendingMembers, err := s.getPendingMembers(org); err == nil {
		for _, m := range pendingMembers {
			members = append(members, m)
		}
	} else {
		return nil, err
	}
	return members, nil
}

func (s *GithubClient) getNonPendingMembers(org string) ([]*Member, error) {
	var query struct {
		Organization struct {
			Login           githubv4.String
			MembersWithRole struct {
				PageInfo PageInfo
				Edges    []struct {
					Node struct {
						Login     githubv4.String
						Name      githubv4.String
						Email     githubv4.String
						CreatedAt githubv4.DateTime
					}
					HasTwoFactorEnabled githubv4.Boolean
					Role                githubv4.String
				}
			} `graphql:"membersWithRole(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}
	var members []*Member
	return members, s.Query(&query).Str("org", org).Cursor("cursor").RunPaginated(func() PageInfo {
		for _, edge := range query.Organization.MembersWithRole.Edges {
			m := &Member{
				Login:               string(edge.Node.Login),
				Name:                string(edge.Node.Name),
				Email:               string(edge.Node.Email),
				HasTwoFactorEnabled: bool(edge.HasTwoFactorEnabled),
				Role:                string(edge.Role),
				CreatedAt:           edge.Node.CreatedAt.Time,
				Pending:             false,
			}
			members = append(members, m)
		}
		return query.Organization.MembersWithRole.PageInfo
	})
}

func (s *GithubClient) getPendingMembers(org string) ([]*Member, error) {
	var query struct {
		Organization struct {
			Login          githubv4.String
			PendingMembers struct {
				PageInfo PageInfo
				Nodes    []struct {
					Login     githubv4.String
					Name      githubv4.String
					Email     githubv4.String
					CreatedAt githubv4.DateTime
				} `graphql:"nodes"`
			} `graphql:"pendingMembers(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}
	var members []*Member
	return members, s.Query(&query).Str("org", org).Cursor("cursor").RunPaginated(func() PageInfo {
		for _, node := range query.Organization.PendingMembers.Nodes {
			m := &Member{
				Login:     string(node.Login),
				Name:      string(node.Name),
				Email:     string(node.Email),
				CreatedAt: node.CreatedAt.Time,
				Pending:   true,
			}
			members = append(members, m)
		}
		return query.Organization.PendingMembers.PageInfo
	})
}

type Team struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Slug            string            `json:"slug,omitempty"`
	CombinedSlug    string            `json:"combined_slug,omitempty"`
	Parent          string            `json:"parent,omitempty"`
	MemberCount     int               `json:"member_count,omitempty"`
	Members         []*TeamMember     `json:"members,omitempty"`
	RepositoryCount int               `json:"repository_count,omitempty"`
	Repositories    []*TeamRepository `json:"repositories,omitempty"`
	ChildCount      int               `json:"child_count,omitempty"`
}

func (s *GithubClient) GetTeams(org string) ([]*Team, error) {
	var query struct {
		Organization struct {
			Teams struct {
				PageInfo PageInfo
				Nodes    []struct {
					ID           githubv4.String
					Name         githubv4.String
					Slug         githubv4.String
					CombinedSlug githubv4.String
					ParentTeam   struct {
						ID githubv4.String
					}
					Members struct {
						TotalCount githubv4.Int
					}
					Repositories struct {
						TotalCount githubv4.Int
					}
					ChildTeams struct {
						TotalCount githubv4.Int
					}
				}
			} `graphql:"teams(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}
	var teams []*Team
	return teams, s.Query(&query).Str("org", org).Cursor("cursor").RunPaginated(func() PageInfo {
		for _, node := range query.Organization.Teams.Nodes {
			t := &Team{
				ID:              string(node.ID),
				Name:            string(node.Name),
				Slug:            string(node.Slug),
				CombinedSlug:    string(node.CombinedSlug),
				Parent:          string(node.ParentTeam.ID),
				MemberCount:     int(node.Members.TotalCount),
				RepositoryCount: int(node.Repositories.TotalCount),
				ChildCount:      int(node.ChildTeams.TotalCount),
			}
			teams = append(teams, t)
		}
		return query.Organization.Teams.PageInfo
	})
}

type TeamMember struct {
	Login string `json:"login,omitempty"`
	Role  string `json:"role,omitempty"`
}

func (s *GithubClient) LoadTeamMembers(org string, teams ...*Team) error {
	for _, team := range teams {
		if err := s.loadTeamMembers(org, team); err != nil {
			return err
		}
	}
	return nil
}

func (s *GithubClient) loadTeamMembers(org string, team *Team) error {
	var query struct {
		Organization struct {
			Team struct {
				Members struct {
					PageInfo PageInfo
					Edges    []struct {
						Node struct {
							Login githubv4.String
						}
						Role githubv4.String
					}
				} `graphql:"members(first: 100, after: $cursor)"`
			} `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $org)"`
	}
	return s.Query(&query).Str("org", org).Str("slug", team.Slug).Cursor("cursor").RunPaginated(func() PageInfo {
		for _, m := range query.Organization.Team.Members.Edges {
			teamMember := &TeamMember{
				Login: string(m.Node.Login),
				Role:  string(m.Role),
			}
			team.Members = append(team.Members, teamMember)
		}
		return query.Organization.Team.Members.PageInfo
	})
}

type TeamRepository struct {
	Owner      string `json:"owner,omitempty"`
	Name       string `json:"name,omitempty"`
	Permission string `json:"permission,omitempty"`
}

func (s *GithubClient) LoadTeamRepositories(org string, teams ...*Team) error {
	for _, team := range teams {
		if err := s.loadTeamRepositories(org, team); err != nil {
			return err
		}
	}
	return nil
}

func (s *GithubClient) loadTeamRepositories(org string, team *Team) error {
	var query struct {
		Organization struct {
			Team struct {
				Repositories struct {
					PageInfo PageInfo
					Edges    []struct {
						Node struct {
							Owner struct {
								Login githubv4.String
							}
							Name githubv4.String
						}
						Permission githubv4.String
					}
				} `graphql:"repositories(first: 100, after: $cursor)"`
			} `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $org)"`
	}
	return s.Query(&query).Str("org", org).Str("slug", team.Slug).Cursor("cursor").RunPaginated(func() PageInfo {
		for _, m := range query.Organization.Team.Repositories.Edges {
			teamRepository := &TeamRepository{
				Owner:      string(m.Node.Owner.Login),
				Name:       string(m.Node.Name),
				Permission: string(m.Permission),
			}
			team.Repositories = append(team.Repositories, teamRepository)
		}
		return query.Organization.Team.Repositories.PageInfo
	})
}

func (s *GithubClient) GetOrganizationRepositories(org string) ([]*Repository, error) {
	var query struct {
		Organization struct {
			Repositories struct {
				PageInfo PageInfo
				Nodes    []struct {
					Owner struct {
						Login githubv4.String
					}
					Name             githubv4.String
					CreatedAt        githubv4.DateTime
					DefaultBranchRef struct {
						Name githubv4.String
					}
					Description          githubv4.String
					DescriptionHTML      githubv4.String `graphql:"descriptionHTML"`
					ShortDescriptionHTML githubv4.String `graphql:"shortDescriptionHTML"`
					HasIssuesEnabled     githubv4.Boolean
					HasProjectsEnabled   githubv4.Boolean
					HasWikiEnabled       githubv4.Boolean
					HomepageURL          githubv4.String
					URL                  githubv4.String
					SshURL               githubv4.String
					IsArchived           githubv4.Boolean
					IsDisabled           githubv4.Boolean
					IsEmpty              githubv4.Boolean
					IsFork               githubv4.Boolean
					IsLocked             githubv4.Boolean
					IsMirror             githubv4.Boolean
					IsPrivate            githubv4.Boolean
					IsTemplate           githubv4.Boolean
					LicenseInfo          struct {
						Name githubv4.String
					}
					Labels struct {
						TotalCount githubv4.Int
					}
					Languages struct {
						TotalCount githubv4.Int
					}
					RebaseMergeAllowed  githubv4.Boolean
					MergeCommitAllowed  githubv4.Boolean
					SquashMergeAllowed  githubv4.Boolean
					DeleteBranchOnMerge githubv4.Boolean
					DiskUsage           githubv4.Int
					PushedAt            githubv4.DateTime
					UpdatedAt           githubv4.DateTime
					PrimaryLanguage     struct {
						Name githubv4.String
					}
				}
			} `graphql:"repositories(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}
	var repositories []*Repository
	return repositories, s.Query(&query).Str("org", org).Cursor("cursor").Run(func() PageInfo {
		for _, node := range query.Organization.Repositories.Nodes {
			r := &Repository{
				Owner:               string(node.Owner.Login),
				Name:                string(node.Name),
				Private:             bool(node.IsPrivate),
				Description:         string(node.Description),
				Homepage:            string(node.HomepageURL),
				URL:                 string(node.URL),
				Ssh:                 string(node.SshURL),
				DefaultBranch:       string(node.DefaultBranchRef.Name),
				Archived:            bool(node.IsArchived),
				Disabled:            bool(node.IsDisabled),
				Template:            bool(node.IsTemplate),
				IssuesEnabled:       bool(node.HasIssuesEnabled),
				ProjectsEnabled:     bool(node.HasProjectsEnabled),
				WikiEnabled:         bool(node.HasWikiEnabled),
				AllowRebaseMerge:    bool(node.RebaseMergeAllowed),
				AllowSquashMerge:    bool(node.SquashMergeAllowed),
				AllowMergeCommit:    bool(node.MergeCommitAllowed),
				DeleteBranchOnMerge: bool(node.DeleteBranchOnMerge),
				LicenseName:         string(node.LicenseInfo.Name),
				Topics:              nil,
				DiskUsage:           int(node.DiskUsage),
				PushedAt:            node.PushedAt.Time,
				UpdatedAt:           node.UpdatedAt.Time,
				Age:                 int(time.Now().Sub(node.PushedAt.Time).Seconds()),
				PrimaryLanguage:     string(node.PrimaryLanguage.Name),
			}
			repositories = append(repositories, r)
		}
		return query.Organization.Repositories.PageInfo
	})
}
