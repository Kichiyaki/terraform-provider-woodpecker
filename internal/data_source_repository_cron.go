package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type repositoryCronDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*repositoryCronDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*repositoryCronDataSource)(nil)

func newRepositoryCronDataSource() datasource.DataSource {
	return &repositoryCronDataSource{}
}

func (d *repositoryCronDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_cron"
}

func (d *repositoryCronDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a cron job in a specific repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "the id of the cron job",
			},
			"repository_id": schema.Int64Attribute{
				Required:    true,
				Description: "the ID of the repository",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "the name of the cron job",
			},
			"schedule": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "[cron expression](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format)",
			},
			"branch": schema.StringAttribute{
				Computed:    true,
				Description: "the name of the branch (uses default branch if empty)",
			},
			"creator_id": schema.Int64Attribute{
				Computed:    true,
				Description: "id of user who created the cron job",
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "date the cron job was created",
			},
		},
	}
}

func (d *repositoryCronDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryCronDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoryCronModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cron, err := d.client.CronGet(data.RepositoryID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get cron job data", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, cron)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
