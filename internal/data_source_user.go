package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type userDataSource struct {
	client woodpecker.Client
}

var _ datasource.DataSource = (*userDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*userDataSource)(nil)

func newUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the user's id",
			},
			"login": schema.StringAttribute{
				Required:    true,
				Description: "The user's login. Use an empty string \"\" to retrieve information about the currently authenticated user.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "the user's email",
			},
			"avatar_url": schema.StringAttribute{
				Computed:    true,
				Description: "the user's avatar URL",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "whether user is active in the system",
			},
			"is_admin": schema.BoolAttribute{
				Computed:    true,
				Description: "whether user is an admin",
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user *woodpecker.User
	var err error
	if login := data.Login.ValueString(); login != "" {
		user, err = d.client.User(login)
	} else {
		user, err = d.client.Self()
	}
	if err != nil {
		resp.Diagnostics.AddError("Couldn't read user data", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, user)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
