package main

import (
	"github.com/engage-wf/core"
	"github.com/engage-wf/github"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	app     = &cobra.Command{
		Use:   "github",
		Short: "Handle Github from engage workflows",
	}
	cmdOrganizations = &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"o", "org"},
		Short:   "Handle organizations on Github",
	}
	cmdRepositories = &cobra.Command{
		Use:     "repositories",
		Aliases: []string{"r", "repos"},
		Short:   "Handle Repositories",
	}
	cmdRepositoriesList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List Repositories",
		Run:     executeRepositoriesList,
	}
	cmdRepositoriesCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a new repository",
		Run:     executeRepositoriesCreate,
	}
	githubClient *github.GithubClient
)

func init() {
	core.DefaultCLI(app, version, commit, "ENGAGE_GITHUB")
	app.PersistentFlags().String("organization", "", "Name of the Organization to operate on")
	cmdRepositories.AddCommand(cmdRepositoriesList, cmdRepositoriesCreate)
	app.AddCommand(cmdOrganizations, cmdRepositories)
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeRepositoriesList(cmd *cobra.Command, args []string) {
	if cmd.Flags().Changed("organization") {
		org, _ := cmd.Flags().GetString("organization")
		result, err := gh().ListOrgRepositories(org)
		if err != nil {
			panic(err)
		}
		core.PrintJSON(result)
	}
}

func executeRepositoriesCreate(cmd *cobra.Command, args []string) {
	var repo github.Repository
	if err := core.ReadFromStdin(&repo); err != nil {
		panic(err)
	}
	org, _ := cmd.Flags().GetString("organization")
	if err := gh().CreateRepository(org, repo); err != nil {
		panic(err)
	}
}

func gh() *github.GithubClient {
	if githubClient == nil {
		githubClient = github.New(viper.GetString("token"))
	}
	return githubClient
}
