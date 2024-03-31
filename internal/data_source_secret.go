package internal

import (
	"context"
	"fmt"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*secretDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*secretDataSource)(nil)

func newSecretDataSource() datasource.DataSource {
	return &secretDataSource{}
}

func (d *secretDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *secretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a global secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the secret's id",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "the name of the secret",
			},
			"events": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "events for which the secret is available " +
					"(push, tag, pull_request, pull_request_closed, deployment, cron, manual, release)",
			},
			"images": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "list of Docker images for which this secret is available",
			},
		},
	}
}

func (d *secretDataSource) Configure(
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

func (d *secretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data secretDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := d.client.GlobalSecret(data.Name.ValueString())
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
