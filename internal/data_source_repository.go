package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type repositoryDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*repositoryDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*repositoryDataSource)(nil)

func newRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the repository's id",
			},
			"forge_remote_id": schema.StringAttribute{
				Computed:    true,
				Description: "the unique identifier for the repository on the forge",
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "the owner of the repository",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "the name of the repository",
			},
			"full_name": schema.StringAttribute{
				Required:    true,
				Description: "the full name of the repository (format: owner/reponame)",
			},
			"avatar_url": schema.StringAttribute{
				Computed:    true,
				Description: "the repository's avatar URL",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "the URL of the repository on the forge",
			},
			"clone_url": schema.StringAttribute{
				Computed:    true,
				Description: "the URL to clone repository",
			},
			"default_branch": schema.StringAttribute{
				Computed:    true,
				Description: "the name of the default branch",
			},
			"scm": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "type of repository " +
					"(see [the source code](https://github.com/woodpecker-ci/woodpecker/blob/main/server/model/const.go#L67))",
			},
			"timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "after this timeout a pipeline has to finish or will be treated as timed out (in minutes)",
			},
			"visibility": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "project visibility (public, private, internal), " +
					"see [the docs](https://woodpecker-ci.org/docs/usage/project-settings#project-visibility) for more info",
			},
			"is_private": schema.BoolAttribute{
				Computed:    true,
				Description: "whether the repo (SCM) is private",
			},
			"is_trusted": schema.BoolAttribute{
				Computed:    true,
				Description: "when true, underlying pipeline containers get access to escalated capabilities like mounting volumes",
			},
			"is_gated": schema.BoolAttribute{
				Computed:    true,
				Description: "when true, every pipeline needs to be approved before being executed",
			},
			"allow_pull_requests": schema.BoolAttribute{
				Computed:    true,
				Description: "Enables handling webhook's pull request event. If disabled, then pipeline won't run for pull requests.",
			},
			"config_file": schema.StringAttribute{
				Computed: true,
				Description: "The path to the pipeline config file or folder. " +
					"By default it is left empty which will use the following configuration " +
					"resolution .woodpecker/*.yml -> .woodpecker/*.yaml -> .woodpecker.yml -> .woodpecker.yaml.",
			},
			"netrc_only_trusted": schema.BoolAttribute{
				Computed: true,
				MarkdownDescription: "whether netrc credentials should be only injected into trusted containers, " +
					"see [the docs](https://woodpecker-ci.org/docs/usage/project-settings#only-inject-netrc-credentials-into-trusted-containers) for more info",
			},
		},
	}
}

func (d *repositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(woodpecker.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected woodpecker.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *repositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoryModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := d.client.RepoLookup(data.FullName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get repository data", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
