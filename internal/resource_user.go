package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type userResource struct {
	client woodpecker.Client
}

var _ resource.Resource = (*userResource)(nil)
var _ resource.ResourceWithConfigure = (*userResource)(nil)
var _ resource.ResourceWithImportState = (*userResource)(nil)

func newUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a user resource.


This resource allows you to add/remove users. When applied, a new user will be created. When destroyed, that user will be removed.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the user's id",
			},
			"login": schema.StringAttribute{
				Required:    true,
				Description: "the name of the user",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "the email of the user",
			},
			"avatar": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "the user's avatar URL",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "whether user is active in the system",
			},
			"admin": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "whether user is an admin",
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.UserPost(wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't create user", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, user)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.User(data.Login.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get user", err.Error())
	}

	resp.Diagnostics.Append(data.setValues(ctx, user)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.UserPatch(wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update user", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, user)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UserDel(data.Login.ValueString()); err != nil {
		resp.Diagnostics.AddError("Couldn't delete user", err.Error())
		return
	}

	// If execution completes without error, the framework will automatically
	// call DeleteResponse.State.RemoveResource(), so it can be omitted
	// from provider logic.
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("login"), req.ID)...)
}
