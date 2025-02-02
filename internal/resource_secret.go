package internal

import (
	"context"
	"fmt"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretResource struct {
	client woodpecker.Client
}

var _ resource.Resource = (*secretResource)(nil)
var _ resource.ResourceWithConfigure = (*secretResource)(nil)
var _ resource.ResourceWithImportState = (*secretResource)(nil)
var _ resource.ResourceWithUpgradeState = (*secretResource)(nil)

func newSecretResource() resource.Resource {
	return &secretResource{}
}

func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allows you to add/remove global secrets." +
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
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							woodpecker.EventPush,
							woodpecker.EventTag,
							woodpecker.EventPull,
							woodpecker.EventPullClosed,
							woodpecker.EventDeploy,
							woodpecker.EventCron,
							woodpecker.EventManual,
							woodpecker.EventRelease,
						),
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

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data secretResourceModelV1

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GlobalSecretCreate(wData)
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

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data secretResourceModelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GlobalSecret(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get secret", err.Error())
	}

	resp.Diagnostics.Append(data.setValues(ctx, secret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data secretResourceModelV1

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GlobalSecretUpdate(wData)
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

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data secretResourceModelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.GlobalSecretDelete(data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Couldn't delete secret", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *secretResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *secretResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
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
						Description: "events for which the secret is available " +
							"(push, tag, pull_request, pull_request_closed, deployment, cron, manual, release)",
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(
									"push",
									"tag",
									"pull_request",
									"pull_request_closed",
									"deployment",
									"cron",
									"manual",
									"release",
								),
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
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData secretResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, secretResourceModelV1{
					ID:     priorStateData.ID,
					Name:   priorStateData.Name,
					Value:  priorStateData.Value,
					Images: priorStateData.Images,
					Events: priorStateData.Events,
				})...)
			},
		},
	}
}
