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
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
)

type repositoryCronResource struct {
	client woodpecker.Client
}

var _ resource.Resource = (*repositoryCronResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryCronResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryCronResource)(nil)

func newRepositoryCronResource() resource.Resource {
	return &repositoryCronResource{}
}

func (r *repositoryCronResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository_cron"
}

func (r *repositoryCronResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allows you to add/remove cron jobs for specific repositories." +
			" When applied, a new cron job will be created." +
			" When destroyed, that cron job will be removed." +
			" For more information see [the Woodpecker docs](https://woodpecker-ci.org/docs/usage/cron).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the id of the cron job",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "the name of the cron job",
			},
			"schedule": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "[cron expression](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format)",
			},
			"branch": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "the name of the branch (uses default branch if empty)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creator_id": schema.Int64Attribute{
				Computed:    true,
				Description: "id of user who created the cron job",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "date the cron job was created",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *repositoryCronResource) Configure(
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

func (r *repositoryCronResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data repositoryCronModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cron, err := r.client.CronCreate(data.RepositoryID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't create cron job", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, cron)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryCronResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryCronModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cron, err := r.client.CronGet(data.RepositoryID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get cron job", err.Error())
	}

	resp.Diagnostics.Append(data.setValues(ctx, cron)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryCronResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data repositoryCronModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cron, err := r.client.CronUpdate(data.RepositoryID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update cron job", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, cron)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryCronResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data repositoryCronModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CronDelete(data.RepositoryID.ValueInt64(), data.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Couldn't delete cron job", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *repositoryCronResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, importStateIDSeparator)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: repository_id/id. Got: %q", req.ID),
		)
		return
	}

	repoID, err := strconv.ParseInt(idParts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository id", err.Error())
		return
	}

	id, err := strconv.ParseInt(idParts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cron id", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository_id"), repoID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
