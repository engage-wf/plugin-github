# Github Plugin for Engage

This plugin allows the use of Github APIs (using both v3 and v4 where appropriate) from within Engage Workflows or as a 
standalone binary.

The latest binary distribution can be downloaded from [Releases](https://github.com/engage-wf/plugin-github/releases/latest)

## Standalone Usage

`github help` shows detailed usage instructions. 

Most commands need authorization against the Github API, which is facilitated by using a personal access token, 
that can be given by argument (`--token <API_TOKEN>`) or from the environment (`ENGAGE_GITHUB_TOKEN=<API_TOKEN>`).

Please note that many commands either need or are more useful with WRITE or ADMIN permissions on the respective objects.

### Building locally

To build the Plugin, execute `make build` in the root of the repository. The binary will be built into the `bin` 
directory.

## Examples

**List all repositories created within an organization**

`github organizations -o $ORGANIZATION_NAME repositories list | jq '[ .[] | .name ]'`

**List all members of an organization together with their role**

`github organizations -o $ORGANIZATION_NAME members list | jq '[ .[] | { name: .login, role: .role } ]'`

**List all repositories that have workflows (Github Actions) defined**

`bin/github organizations -o $ORGANIZATION_NAME repositories list --workflows | jq '[ .[] | select(.workflows) | .name ]'`


