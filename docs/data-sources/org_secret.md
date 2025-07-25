---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "woodpecker_org_secret Data Source - terraform-provider-woodpecker"
subcategory: ""
description: |-
  Use this data source to retrieve information about a secret in a specific organization.
---

# woodpecker_org_secret (Data Source)

Use this data source to retrieve information about a secret in a specific organization.

## Example Usage

```terraform
data "woodpecker_org_secret" "test" {
  org_id = 111
  name   = "test"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) the name of the secret
- `org_id` (Number) the ID of the organization

### Read-Only

- `events` (Set of String) events for which the secret is available (push, tag, pull_request, pull_request_closed, deployment, cron, manual, release)
- `id` (Number) the secret's id
- `images` (Set of String) list of Docker images for which this secret is available
