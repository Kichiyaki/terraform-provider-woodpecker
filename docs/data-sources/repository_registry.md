---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "woodpecker_repository_registry Data Source - terraform-provider-woodpecker"
subcategory: ""
description: |-
  Use this data source to retrieve information about a container registry in a specific repository.
---

# woodpecker_repository_registry (Data Source)

Use this data source to retrieve information about a container registry in a specific repository.

## Example Usage

```terraform
resource "woodpecker_repository" "test_repo" {
  full_name = "Kichiyaki/test-repo"
}

data "woodpecker_repository_secret" "test_secret" {
  repository_id = woodpecker_repository.test_repo.id
  address       = "docker.io"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address` (String) the address of the registry (e.g. docker.io)
- `repository_id` (Number) the ID of the repository

### Read-Only

- `email` (String) email used for authentication
- `id` (Number) the id of the registry
- `username` (String) username used for authentication