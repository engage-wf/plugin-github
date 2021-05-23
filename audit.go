package github

import log "github.com/mtrense/soil/logging"

type FullAudit struct {
	Members      []*Member     `json:"members,omitempty"`
	Repositories []*Repository `json:"repositories,omitempty"`
}

type Permission struct {
	RepositoryOwner string `json:"repository_owner,omitempty"`
	RepositoryName  string `json:"repository_name,omitempty"`
	Source          string `json:"source,omitempty"`
	Role            string `json:"role,omitempty"`
}

func (s *GithubClient) FullAudit(org string) (FullAudit, error) {
	var audit FullAudit
	var err error
	if audit.Members, err = s.GetMembers(org); err != nil {
		return audit, err
	}
	if audit.Repositories, err = s.GetOrganizationRepositories(org); err != nil {
		return audit, err
	}
	if err := s.LoadRepositoryCollaborators(audit.Repositories...); err != nil {
		return audit, err
	}
	return audit, nil
}

type TeamMembershipAudit struct {
	Login       string           `json:"login,omitempty"`
	Memberships []TeamMembership `json:"memberships,omitempty"`
}

type TeamMembership struct {
	TeamName string `json:"team_name,omitempty"`
	Role     string `json:"role,omitempty"`
}

func (s *GithubClient) TeamMembershipAudit(org string) ([]TeamMembershipAudit, error) {
	memberships := make(map[string][]TeamMembership)
	var audit []TeamMembershipAudit
	if teams, err := s.GetTeams(org); err == nil {
		if err := s.LoadTeamMembers(org, teams...); err != nil {
			return audit, err
		}
		for _, team := range teams {
			for _, m := range team.Members {
				memberships[m.Login] = append(memberships[m.Login], TeamMembership{
					TeamName: team.Name,
					Role:     m.Role,
				})
			}
		}
		for login, ms := range memberships {
			audit = append(audit, TeamMembershipAudit{
				Login:       login,
				Memberships: ms,
			})
		}
		return audit, nil
	} else {
		return audit, err
	}
}

type TeamPermissionAudit struct {
	Name        string       `json:"name,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

func (s *GithubClient) TeamPermissionAudit(org string) ([]TeamPermissionAudit, error) {
	var audit []TeamPermissionAudit
	if teams, err := s.GetTeams(org); err == nil {
		if err := s.LoadTeamRepositories(org, teams...); err != nil {
			return audit, err
		}
		for _, team := range teams {
			a := TeamPermissionAudit{
				Name: team.Name,
			}
			for _, repo := range team.Repositories {
				a.Permissions = append(a.Permissions, Permission{
					RepositoryOwner: repo.Owner,
					RepositoryName:  repo.Name,
					Role:            repo.Permission,
				})
			}
			audit = append(audit, a)
		}
		return audit, nil
	} else {
		return audit, err
	}
}

type MemberPermissionAudit struct {
	Login       string       `json:"login,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

func (s *GithubClient) MemberPermissionAudit(org string) ([]MemberPermissionAudit, error) {
	memberships := make(map[string][]Permission)
	var audit []MemberPermissionAudit
	log.L().Info().Msg("Fetching Repositories")
	if repositories, err := s.GetOrganizationRepositories(org); err == nil {
		log.L().Info().Msg("Fetching Repository Collaborators")
		if err := s.LoadRepositoryCollaborators(repositories...); err != nil {
			return audit, err
		}
		for _, repository := range repositories {
			for _, collaborator := range repository.Collaborators {
				for _, source := range collaborator.Sources {
					if source.SourceType == "Repository" {
						memberships[collaborator.Login] = append(memberships[collaborator.Login], Permission{
							RepositoryOwner: repository.Owner,
							RepositoryName:  repository.Name,
							Source:          source.SourceType,
							Role:            source.Permission,
						})
					}
				}
			}
		}
		for login, perms := range memberships {
			audit = append(audit, MemberPermissionAudit{
				Login:       login,
				Permissions: perms,
			})
		}
		return audit, nil
	} else {
		return audit, err
	}
}

type ActionsAudit struct {
	Actions            []*ActionAudit `json:"actions,omitempty"`
	TotalUsageWeighted int64          `json:"total_usage_weighted,omitempty"`
}

type ActionAudit struct {
	Repository           string  `json:"repository,omitempty"`
	WorkflowName         string  `json:"workflow_name,omitempty"`
	UsageWeighted        int64   `json:"usage_summary_weighted,omitempty"`
	FractionOfTotalUsage float64 `json:"fraction_of_total_usage,omitempty"`
	UsageMac             int64   `json:"usage_mac,omitempty"`
	UsageUbuntu          int64   `json:"usage_ubuntu,omitempty"`
	UsageWindows         int64   `json:"usage_windows,omitempty"`
}

func (s *GithubClient) ActionsAudit(org string) (ActionsAudit, error) {
	var audit ActionsAudit
	if repositories, err := s.GetOrganizationRepositories(org); err == nil {
		if err := s.LoadRepositoryWorkflows(repositories...); err != nil {
			return audit, err
		}
		var totalUsage int64
		for _, repository := range repositories {
			for _, workflow := range repository.Workflows {
				usageWeighted := workflow.UsageUbuntu + workflow.UsageWindows*2 + workflow.UsageMac*10
				aa := ActionAudit{
					Repository:    repository.Name,
					WorkflowName:  workflow.Name,
					UsageWeighted: usageWeighted,
					UsageMac:      workflow.UsageMac,
					UsageUbuntu:   workflow.UsageUbuntu,
					UsageWindows:  workflow.UsageWindows,
				}
				audit.Actions = append(audit.Actions, &aa)
				totalUsage += usageWeighted
			}
		}
		audit.TotalUsageWeighted = totalUsage
		for _, a := range audit.Actions {
			a.FractionOfTotalUsage = float64(a.UsageWeighted) / float64(totalUsage)
		}
		return audit, nil
	} else {
		return audit, err
	}
}
