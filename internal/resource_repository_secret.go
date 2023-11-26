package internal

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositorySecretResource struct {
	client woodpecker.Client
}

var _ resource.Resource = (*repositorySecretResource)(nil)
var _ resource.ResourceWithConfigure = (*repositorySecretResource)(nil)
var _ resource.ResourceWithImportState = (*repositorySecretResource)(nil)
var _ resource.ResourceWithUpgradeState = (*repositorySecretResource)(nil)

func newRepositorySecretResource() resource.Resource {
	return &repositorySecretResource{}
}

func (r *repositorySecretResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository_secret"
}

func (r *repositorySecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allows you to add/remove secrets that" +
			" are only available to specific repositories." +
			" When applied, a new secret will be created." +
			" When destroyed, that secret will be removed." +
			" For more information see [the Woodpecker docs](https://woodpecker-ci.org/docs/usage/secrets).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the secret's id",
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
				Description: "the name of the secret",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "the value of the secret",
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"events": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "events for which the secret is available (push, tag, pull_request, deployment, cron, manual)",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOfCaseInsensitive("push", "tag", "pull_request", "deployment", "cron", "manual"),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"images": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "list of Docker images for which this secret is available, leave blank to allow all images",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Version: 1,
	}
}

func (r *repositorySecretResource) Configure(
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

func (r *repositorySecretResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data repositorySecretResourceModelV1

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.SecretCreate(data.RepositoryID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't create secret", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, secret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositorySecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositorySecretResourceModelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.Secret(data.RepositoryID.ValueInt64(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get secret", err.Error())
	}

	resp.Diagnostics.Append(data.setValues(ctx, secret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositorySecretResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data repositorySecretResourceModelV1

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.SecretUpdate(data.RepositoryID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update secret", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, secret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositorySecretResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data repositorySecretResourceModelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SecretDelete(data.RepositoryID.ValueInt64(), data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Couldn't delete secret", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *repositorySecretResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, importStateIDSeparator)

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: repository_id/name. Got: %q", req.ID),
		)
		return
	}

	repoID, err := strconv.ParseInt(idParts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository id", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository_id"), repoID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
}

func (r *repositorySecretResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: {
			// remove plugins_only attribute
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:    true,
						Description: "the secret's id",
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
						Description: "the name of the secret",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"value": schema.StringAttribute{
						Required:    true,
						Description: "the value of the secret",
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"events": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "events for which the secret is available (push, tag, pull_request, deployment, cron, manual)",
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.OneOfCaseInsensitive("push", "tag", "pull_request", "deployment", "cron", "manual"),
							),
						},
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
					},
					"plugins_only": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						MarkdownDescription: "whether secret is only available for" +
							" [plugins](https://woodpecker-ci.org/docs/usage/plugins/plugins)",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"images": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Description: "list of Docker images for which this secret is available, leave blank to allow all images",
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData repositorySecretResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, repositorySecretResourceModelV1{
					ID:           priorStateData.ID,
					RepositoryID: priorStateData.RepositoryID,
					Name:         priorStateData.Name,
					Value:        priorStateData.Value,
					Images:       priorStateData.Images,
					Events:       priorStateData.Events,
				})...)
			},
		},
	}
}
