package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type repositoryRegistryDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*repositoryRegistryDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*repositoryRegistryDataSource)(nil)

func newRepositoryRegistryDataSource() datasource.DataSource {
	return &repositoryRegistryDataSource{}
}

func (d *repositoryRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_registry"
}

func (d *repositoryRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a container registry in a specific repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the id of the registry",
			},
			"repository_id": schema.Int64Attribute{
				Required:    true,
				Description: "the ID of the repository",
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "the address of the registry (e.g. docker.io)",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "username used for authentication",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "email used for authentication",
			},
		},
	}
}

func (d *repositoryRegistryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoryRegistryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := d.client.Registry(data.RepositoryID.ValueInt64(), data.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get registry data", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, registry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
