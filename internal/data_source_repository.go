package internal

import (
	"context"
	"fmt"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*repositoryDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*repositoryDataSource)(nil)

func newRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

func (d *repositoryDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the repository's id",
			},
			"forge_id": schema.Int64Attribute{
				Computed:    true,
				Description: "the forge's id",
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
			"forge_url": schema.StringAttribute{
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
			"timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "after this timeout a pipeline has to finish or will be treated as timed out (in minutes)",
			},
			"visibility": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: fmt.Sprintf(
					"project visibility (%s, %s, %s), ",
					woodpecker.VisibilityModePublic.String(),
					woodpecker.VisibilityModePrivate.String(),
					woodpecker.VisibilityModeInternal.String(),
				) +
					"see [the docs](https://woodpecker-ci.org/docs/usage/project-settings#project-visibility) for more info",
			},
			"is_private": schema.BoolAttribute{
				Computed:    true,
				Description: "whether the repo (SCM) is private",
			},
			"trusted": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"network": schema.BoolAttribute{
						Computed:    true,
						Description: "Pipeline containers get access to network privileges like changing DNS.",
					},
					"security": schema.BoolAttribute{
						Computed:    true,
						Description: "Pipeline containers get access to security privileges.",
					},
					"volumes": schema.BoolAttribute{
						Computed:    true,
						Description: "Pipeline containers are allowed to mount volumes.",
					},
				},
				Computed: true,
			},
			"require_approval": schema.StringAttribute{
				Computed: true,
				Description: "Prevent malicious pipelines from exposing secrets or " +
					"running harmful tasks by approving them before execution. " +
					fmt.Sprintf(
						"Allowed values: %s, %s, %s",
						woodpecker.ApprovalModeForks.String(),
						woodpecker.ApprovalModePullRequests.String(),
						woodpecker.ApprovalModeAllEvents.String(),
					),
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "whether the repo is active",
			},
			"allow_pull_requests": schema.BoolAttribute{
				Computed: true,
				Description: "Enables handling webhook's pull request event." +
					" If disabled, then pipeline won't run for pull requests.",
			},
			"allow_deployments": schema.BoolAttribute{
				Computed:    true,
				Description: "Enables a pipeline to be started with the deploy event from a successful pipeline.",
			},
			"config_file": schema.StringAttribute{
				Computed: true,
				Description: "The path to the pipeline config file or folder. " +
					"By default it is left empty which will use the following configuration " +
					"resolution .woodpecker/*.yml -> .woodpecker/*.yaml -> .woodpecker.yml -> .woodpecker.yaml.",
			},
			"cancel_previous_pipeline_events": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Enables to cancel pending and running pipelines of the same " +
					fmt.Sprintf(
						"event and context before starting the newly triggered one (%s, %s, %s, %s).",
						woodpecker.EventPush,
						woodpecker.EventTag,
						woodpecker.EventPull,
						woodpecker.EventDeploy,
					),
			},
			"netrc_trusted_plugins": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Plugins that get access to netrc credentials that can " +
					"be used to clone repositories from the forge or push them into the forge.",
			},
		},
	}
}

func (d *repositoryDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(woodpecker.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected woodpecker.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
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
