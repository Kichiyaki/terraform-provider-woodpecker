package internal

import (
	"context"
	"fmt"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositorySecretDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*repositorySecretDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*repositorySecretDataSource)(nil)

func newRepositorySecretDataSource() datasource.DataSource {
	return &repositorySecretDataSource{}
}

func (d *repositorySecretDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository_secret"
}

func (d *repositorySecretDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a secret in a specific repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the secret's id",
			},
			"repository_id": schema.Int64Attribute{
				Required:    true,
				Description: "the ID of the repository",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "the name of the secret",
			},
			"events": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "events for which the secret is available " +
					fmt.Sprintf(
						"(%s, %s, %s, %s, %s, %s, %s, %s)",
						woodpecker.EventPush,
						woodpecker.EventTag,
						woodpecker.EventPull,
						woodpecker.EventPullClosed,
						woodpecker.EventDeploy,
						woodpecker.EventCron,
						woodpecker.EventManual,
						woodpecker.EventRelease,
					),
			},
			"images": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "list of Docker images for which this secret is available",
			},
		},
	}
}

func (d *repositorySecretDataSource) Configure(
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

func (d *repositorySecretDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data repositorySecretDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := d.client.Secret(data.RepositoryID.ValueInt64(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get secret data", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, secret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
