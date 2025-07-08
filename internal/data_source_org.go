package internal

import (
	"context"
	"fmt"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type orgDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*orgDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*orgDataSource)(nil)

func newOrgDataSource() datasource.DataSource {
	return &orgDataSource{}
}

func (d *orgDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *orgDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the org's id",
			},
			"forge_id": schema.Int64Attribute{
				Computed:    true,
				Description: "the forge's id",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "the org's name",
			},
			"is_user": schema.BoolAttribute{
				Computed:    true,
				Description: "whether org is a user",
			},
		},
	}
}

func (d *orgDataSource) Configure(
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

func (d *orgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orgModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := d.client.OrgLookup(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get org data", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, org)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
