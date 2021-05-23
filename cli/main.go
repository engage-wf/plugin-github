package main

import (
	"github.com/engage-wf/core"
	github "github.com/engage-wf/plugin-github"
	. "github.com/mtrense/soil/config"
	"github.com/mtrense/soil/logging"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("github",
		Short("Handle Github from engage workflows"),
		SubCommand("organizations",
			Short("Handle organizations on Github"),
			Alias("o", "org"),
			Flag("organization", Str(""), Abbr("o"), Description("Organization to handle"), Persistent(), Mandatory(), Env()),
			SubCommand("members",
				Short("Members of this Organization"),
				Alias("m"),
				SubCommand("list",
					Short("List Members of this Organization"),
					Alias("l"),
					Run(executeOrganizationMembersList),
				),
			),
			SubCommand("teams",
				Short("Teams defined in this Organization"),
				Alias("t"),
				SubCommand("list",
					Short("List Teams defined in this Organization"),
					Alias("l"),
					Flag("members", Bool(), Description("Include Members in listing"), Persistent()),
					Run(executeOrganizationTeamsList),
				),
			),
			SubCommand("repositories",
				Short("Repositories in this Organization"),
				Alias("r"),
				SubCommand("list",
					Short("List Repositories in this Organization"),
					Alias("l"),
					Flag("security", Bool(), Description("Fetch Security Configuration"), Persistent()),
					Flag("branch-protection", Bool(), Description("Fetch Branch Protection Configuration"), Persistent()),
					Flag("languages", Bool(), Description("Fetch Repository Languages"), Persistent()),
					Flag("workflows", Bool(), Description("Fetch defined Workflows"), Persistent()),
					Flag("pattern", Str(""), Description("Pattern to match the Repository name against"), Persistent()),
					Run(executeOrganizationRepositoriesList),
				),
			),
			SubCommand("audit",
				Short("Fetch user and permission information for the organization"),
				SubCommand("full",
					Short("Generate a full audit"),
					Alias("f"),
					Run(executeOrganizationAuditFull),
				),
				SubCommand("team-membership",
					Short("Generate an audit on Team Membership"),
					Alias("tm"),
					Run(executeOrganizationAuditTeamMembership),
				),
				SubCommand("team-permission",
					Short("Generate an audit on Team Permissions"),
					Alias("tp"),
					Run(executeOrganizationAuditTeamPermission),
				),
				SubCommand("member-permission",
					Short("Generate an audit on Member Permissions"),
					Alias("mp"),
					Run(executeOrganizationAuditMemberPermission),
				),
				SubCommand("actions",
					Short("Generate an audit on configured Actions"),
					Alias("a"),
					Run(executeOrganizationAuditActions),
				),
			),
		),
		SubCommand("repositories",
			Short("Handle Repositories"),
			Alias("r", "repos", "repository"),
			Flag("user", Str(""), Abbr("u"), Description("Name of the User/Organization to operate on"), Mandatory(), Env(), Persistent()),
			SubCommand("list",
				Short("List Repositories"),
				Alias("l", "ls"),
				//Run(executeRepositoriesList),
			),
			SubCommand("create",
				Short("Create a new Repository"),
				Alias("c"),
				Run(executeRepositoriesCreate),
			),
		),
		SubCommand("user",
			Short("Handle Users"),
			Alias("u", "usr"),
			SubCommand("keys",
				Short("Handle a user's keys"),
				Alias("k"),
				SubCommand("list",
					Short("List keys"),
					Alias("l", "ls"),
					Run(executeUserKeysList),
				),
			),
		),
		Flag("token", Str(""), Description("Authentication token used to authenticate agains the Github API"), Env(), Persistent()),
		FlagLogLevel("warn"),
		FlagLogFormat(),
		FlagLogFile(),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
	githubClient *github.GithubClient
)

func init() {
	EnvironmentConfig("ENGAGE_GITHUB")
	logging.ConfigureDefaultLogging()
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeOrganizationMembersList(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	if members, err := gh().GetMembers(org); err == nil {
		core.PrintJSON(members)
	} else {
		panic(err)
	}
}

func executeOrganizationTeamsList(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	members, _ := cmd.Flags().GetBool("members")
	client := gh()
	if teams, err := client.GetTeams(org); err == nil {
		if members {
			if err := client.LoadTeamMembers(org, teams...); err != nil {
				panic(err)
			}
		}
		core.PrintJSON(teams)
	} else {
		panic(err)
	}
}

func executeOrganizationRepositoriesList(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	security, _ := cmd.Flags().GetBool("security")
	branchProtection, _ := cmd.Flags().GetBool("branch-protection")
	languages, _ := cmd.Flags().GetBool("languages")
	workflows, _ := cmd.Flags().GetBool("workflows")
	client := gh()
	if repositories, err := client.GetOrganizationRepositories(org); err == nil {
		if security {
			if err := client.LoadRepositorySecurityConfig(repositories...); err != nil {
				panic(err)
			}
		}
		if branchProtection {
			if err := client.LoadRepositoryBranchProtectionRules(repositories...); err != nil {
				panic(err)
			}
		}
		if languages {
			if err := client.LoadRepositoryLanguages(repositories...); err != nil {
				panic(err)
			}
		}
		if workflows {
			if err := client.LoadRepositoryWorkflows(repositories...); err != nil {
				panic(err)
			}
		}
		core.PrintJSON(repositories)
	} else {
		panic(err)
	}
}

func executeOrganizationAuditFull(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	if audit, err := gh().FullAudit(org); err == nil {
		core.PrintJSON(audit)
	} else {
		panic(err)
	}
}

func executeOrganizationAuditTeamMembership(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	if audit, err := gh().TeamMembershipAudit(org); err == nil {
		core.PrintJSON(audit)
	} else {
		panic(err)
	}
}

func executeOrganizationAuditTeamPermission(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	if audit, err := gh().TeamPermissionAudit(org); err == nil {
		core.PrintJSON(audit)
	} else {
		panic(err)
	}
}

func executeOrganizationAuditMemberPermission(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	if audit, err := gh().MemberPermissionAudit(org); err == nil {
		core.PrintJSON(audit)
	} else {
		panic(err)
	}
}

func executeOrganizationAuditActions(cmd *cobra.Command, args []string) {
	org, _ := cmd.Flags().GetString("organization")
	if audit, err := gh().ActionsAudit(org); err == nil {
		core.PrintJSON(audit)
	} else {
		panic(err)
	}
}

func executeRepositoriesCreate(cmd *cobra.Command, args []string) {
	var repo github.Repository
	if err := core.ReadFromStdin(&repo); err != nil {
		panic(err)
	}
	org, _ := cmd.Flags().GetString("organization")
	if err := gh().CreateRepository(org, &repo); err != nil {
		panic(err)
	}
}

func executeUserKeysList(cmd *cobra.Command, args []string) {

}

func gh() *github.GithubClient {
	if githubClient == nil {
		githubClient = github.New(viper.GetString("token"))
	}
	return githubClient
}
