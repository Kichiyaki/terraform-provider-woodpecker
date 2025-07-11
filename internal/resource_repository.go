package internal

import (
	"context"
	"fmt"
	"slices"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryResource struct {
	client woodpecker.Client
}

var _ resource.Resource = (*repositoryResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryResource)(nil)

func newRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

func (r *repositoryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Provides a repository resource.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "the repository's id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"forge_id": schema.Int64Attribute{
				Computed:    true,
				Description: "the forge's id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"forge_remote_id": schema.StringAttribute{
				Computed:    true,
				Description: "the unique identifier for the repository on the forge",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"forge_url": schema.StringAttribute{
				Computed:    true,
				Description: "the URL of the repository on the forge",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"clone_url": schema.StringAttribute{
				Computed:    true,
				Description: "the URL to clone repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "the owner of the repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "the name of the repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"full_name": schema.StringAttribute{
				Required:    true,
				Description: "the full name of the repository (format: owner/reponame)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"avatar_url": schema.StringAttribute{
				Computed:    true,
				Description: "the repository's avatar URL",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_branch": schema.StringAttribute{
				Computed:    true,
				Description: "the name of the default branch",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "after this timeout a pipeline has to finish or will be treated as timed out (in minutes)",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: fmt.Sprintf(
					"project visibility (%s, %s, %s), ",
					woodpecker.VisibilityModePublic.String(),
					woodpecker.VisibilityModePrivate.String(),
					woodpecker.VisibilityModeInternal.String(),
				) +
					"see [the docs](https://woodpecker-ci.org/docs/usage/project-settings#project-visibility) for more info",
				Validators: []validator.String{
					stringvalidator.OneOf(
						woodpecker.VisibilityModePublic.String(),
						woodpecker.VisibilityModePrivate.String(),
						woodpecker.VisibilityModeInternal.String(),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_private": schema.BoolAttribute{
				Computed:    true,
				Description: "whether the repo (SCM) is private",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"trusted": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"network": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Pipeline containers get access to network privileges like changing DNS.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"security": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Pipeline containers get access to security privileges.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"volumes": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Pipeline containers are allowed to mount volumes.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"require_approval": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Prevents malicious pipelines from exposing secrets or " +
					"running harmful tasks by approving them before execution. " +
					fmt.Sprintf(
						"Allowed values: %s, %s, %s",
						woodpecker.ApprovalModeForks.String(),
						woodpecker.ApprovalModePullRequests.String(),
						woodpecker.ApprovalModeAllEvents.String(),
					),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						woodpecker.ApprovalModeForks.String(),
						woodpecker.ApprovalModePullRequests.String(),
						woodpecker.ApprovalModeAllEvents.String(),
					),
				},
			},
			"approval_allowed_users": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "the list of users who's pipelines never require an approval",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "whether the repo is active",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_pull_requests": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Enables handling webhook's pull request event." +
					" If disabled, then pipeline won't run for pull requests.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_deployments": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enables a pipeline to be started with the deploy event from a successful pipeline.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"config_file": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The path to the pipeline config file or folder. " +
					"By default, it is left empty which will use the following configuration " +
					"resolution .woodpecker/*.yml -> .woodpecker/*.yaml -> .woodpecker.yml -> .woodpecker.yaml.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cancel_previous_pipeline_events": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Enables to cancel pending and running pipelines of the same " +
					fmt.Sprintf(
						"event and context before starting the newly triggered one (%s, %s, %s, %s).",
						woodpecker.EventPush,
						woodpecker.EventTag,
						woodpecker.EventPull,
						woodpecker.EventDeploy,
					),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							woodpecker.EventPush,
							woodpecker.EventTag,
							woodpecker.EventPull,
							woodpecker.EventDeploy,
						),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"netrc_trusted_plugins": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Plugins that get access to netrc credentials that can " +
					"be used to clone repositories from the forge or push them into the forge.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *repositoryResource) Configure(
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

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoFullName := data.FullName.ValueString()

	repos, err := r.client.RepoList(woodpecker.RepoListOptions{All: true})
	if err != nil {
		resp.Diagnostics.AddError("Couldn't list repositories", err.Error())
		return
	}

	idx := slices.IndexFunc(repos, func(repo *woodpecker.Repo) bool {
		return repo.FullName == repoFullName
	})
	if idx < 0 {
		resp.Diagnostics.AddError("Repository not found", fmt.Sprintf("Repository with name '%s' not found", repoFullName))
		return
	}

	wData, diags := data.toWoodpeckerPatch(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	activatedRepo, err := r.client.RepoPost(woodpecker.RepoPostOptions{
		ForgeRemoteID: repos[idx].ForgeRemoteID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Couldn't activate repository", err.Error())
		return
	}

	updatedRepo, err := r.client.RepoPatch(activatedRepo.ID, wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update repository", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, updatedRepo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.RepoLookup(data.FullName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get repository", err.Error())
	}

	resp.Diagnostics.Append(data.setValues(ctx, repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wData, diags := data.toWoodpeckerPatch(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.RepoPatch(data.ID.ValueInt64(), wData)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update repository", err.Error())
		return
	}

	resp.Diagnostics.Append(data.setValues(ctx, repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RepoDel(data.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Couldn't delete repository", err.Error())
		return
	}

	// If execution completes without error, the framework will automatically
	// call DeleteResponse.State.RemoveResource(), so it can be omitted
	// from provider logic.
}

func (r *repositoryResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("full_name"), req.ID)...)
}
