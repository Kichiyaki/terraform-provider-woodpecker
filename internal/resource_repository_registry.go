package internal

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type repositoryRegistryResource struct {
	client woodpecker.Client
}

var _ resource.Resource = (*repositoryRegistryResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryRegistryResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryRegistryResource)(nil)

func newRepositoryRegistryResource() resource.Resource {
	return &repositoryRegistryResource{}
}

func (r *repositoryRegistryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository_registry"
}

func (r *repositoryRegistryResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allows you to add/remove container registries for specific repositories." +
			" When applied, a new registry will be created." +
			" When destroyed, that registry will be removed." +
			" For more information see [the Woodpecker docs](https://woodpecker-ci.org/docs/usage/registries).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the id of the registry",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"repository_id": schema.Int64Attribute{
				Required:    true,
				Description: "the ID of the repository",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "the address of the registry (e.g. docker.io)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "username used for authentication",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Description: "password used for authentication",
				Sensitive:   true,
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "email used for authentication",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *repositoryRegistryResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
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

	r.client = client
}

func (r *repositoryRegistryResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data repositoryRegistryResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RegistryCreate(data.RepositoryID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't create registry", err.Error())
		return
	}

	// RegistryCreate doesn't return ID
	registry, err := r.client.Registry(data.RepositoryID.ValueInt64(), wData.Address)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get registry", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, registry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryRegistryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.Registry(data.RepositoryID.ValueInt64(), data.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get registry", err.Error())
	}

	resp.Diagnostics.Append(data.setValues(ctx, registry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryRegistryResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data repositoryRegistryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.RegistryUpdate(data.RepositoryID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update registry", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, registry)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryRegistryResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data repositoryRegistryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RegistryDelete(data.RepositoryID.ValueInt64(), data.Address.ValueString()); err != nil {
		resp.Diagnostics.AddError("Couldn't delete registry", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *repositoryRegistryResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, importStateIDSeparator)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: repository_id/address. Got: %q", req.ID),
		)
		return
	}

	repoID, err := strconv.ParseInt(idParts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository id", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository_id"), repoID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), idParts[1])...)
}
