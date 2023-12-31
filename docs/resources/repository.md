---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "woodpecker_repository Resource - terraform-provider-woodpecker"
subcategory: ""
description: |-
  Provides a repository resource.
---

# woodpecker_repository (Resource)

Provides a repository resource.

## Example Usage

```terraform
resource "woodpecker_repository" "test_repo" {
  full_name  = "Kichiyaki/test-repo"
  is_trusted = true
  visibility = "public"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `full_name` (String) the full name of the repository (format: owner/reponame)

### Optional

- `allow_pull_requests` (Boolean) Enables handling webhook's pull request event. If disabled, then pipeline won't run for pull requests.
- `config_file` (String) The path to the pipeline config file or folder. By default it is left empty which will use the following configuration resolution .woodpecker/*.yml -> .woodpecker/*.yaml -> .woodpecker.yml -> .woodpecker.yaml.
- `is_gated` (Boolean) when true, every pipeline needs to be approved before being executed
- `is_trusted` (Boolean) when true, underlying pipeline containers get access to escalated capabilities like mounting volumes
- `netrc_only_trusted` (Boolean) whether netrc credentials should be only injected into trusted containers, see [the docs](https://woodpecker-ci.org/docs/usage/project-settings#only-inject-netrc-credentials-into-trusted-containers) for more info
- `timeout` (Number) after this timeout a pipeline has to finish or will be treated as timed out (in minutes)
- `visibility` (String) project visibility (public, private, internal), see [the docs](https://woodpecker-ci.org/docs/usage/project-settings#project-visibility) for more info

### Read-Only

- `avatar_url` (String) the repository's avatar URL
- `clone_url` (String) the URL to clone repository
- `default_branch` (String) the name of the default branch
- `forge_remote_id` (String) the unique identifier for the repository on the forge
- `id` (Number) the repository's id
- `is_private` (Boolean) whether the repo (SCM) is private
- `name` (String) the name of the repository
- `owner` (String) the owner of the repository
- `scm` (String) type of repository (see [the source code](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/const.go#L67))
- `url` (String) the URL of the repository on the forge

## Import

Import is supported using the following syntax:

```shell
terraform import woodpecker_repository.test "<owner>/<repo>"
```
