package internal

import (
	"context"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type userModel struct {
	ID        types.Int64  `tfsdk:"id"`
	ForgeID   types.Int64  `tfsdk:"forge_id"`
	Login     types.String `tfsdk:"login"`
	Email     types.String `tfsdk:"email"`
	AvatarURL types.String `tfsdk:"avatar_url"`
	IsAdmin   types.Bool   `tfsdk:"is_admin"`
}

func (m *userModel) setValues(_ context.Context, user *woodpecker.User) diag.Diagnostics {
	m.ID = types.Int64Value(user.ID)
	m.ForgeID = types.Int64Value(user.ForgeID)
	m.Login = types.StringValue(user.Login)
	m.Email = types.StringValue(user.Email)
	m.AvatarURL = types.StringValue(user.Avatar)
	m.IsAdmin = types.BoolValue(user.Admin)
	return nil
}

func (m *userModel) toWoodpeckerModel(_ context.Context) (*woodpecker.User, diag.Diagnostics) {
	return &woodpecker.User{
		ID:      m.ID.ValueInt64(),
		ForgeID: m.ForgeID.ValueInt64(),
		Login:   m.Login.ValueString(),
		Email:   m.Email.ValueString(),
		Avatar:  m.AvatarURL.ValueString(),
		Admin:   m.IsAdmin.ValueBool(),
	}, nil
}

type secretResourceModelV0 struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Images      types.Set    `tfsdk:"images"`
	PluginsOnly types.Bool   `tfsdk:"plugins_only"`
	Events      types.Set    `tfsdk:"events"`
}

type secretResourceModelV1 struct {
	ID     types.Int64  `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	Images types.Set    `tfsdk:"images"`
	Events types.Set    `tfsdk:"events"`
}

func (m *secretResourceModelV1) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

func (m *secretResourceModelV1) toWoodpeckerModel(ctx context.Context) (*woodpecker.Secret, diag.Diagnostics) {
	var diags diag.Diagnostics

	secret := &woodpecker.Secret{
		ID:    m.ID.ValueInt64(),
		Name:  m.Name.ValueString(),
		Value: m.Value.ValueString(),
	}
	diags.Append(m.Images.ElementsAs(ctx, &secret.Images, false)...)
	diags.Append(m.Events.ElementsAs(ctx, &secret.Events, false)...)

	return secret, diags
}

type secretDataSourceModel struct {
	ID     types.Int64  `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Images types.Set    `tfsdk:"images"`
	Events types.Set    `tfsdk:"events"`
}

func (m *secretDataSourceModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

type repositoryModel struct {
	ID                           types.Int64  `tfsdk:"id"`
	ForgeID                      types.Int64  `tfsdk:"forge_id"`
	ForgeRemoteID                types.String `tfsdk:"forge_remote_id"`
	Owner                        types.String `tfsdk:"owner"`
	Name                         types.String `tfsdk:"name"`
	FullName                     types.String `tfsdk:"full_name"`
	AvatarURL                    types.String `tfsdk:"avatar_url"`
	ForgeURL                     types.String `tfsdk:"forge_url"`
	CloneURL                     types.String `tfsdk:"clone_url"`
	DefaultBranch                types.String `tfsdk:"default_branch"`
	Timeout                      types.Int64  `tfsdk:"timeout"`
	Visibility                   types.String `tfsdk:"visibility"`
	IsPrivate                    types.Bool   `tfsdk:"is_private"`
	Trusted                      types.Object `tfsdk:"trusted"`
	RequireApproval              types.String `tfsdk:"require_approval"`
	IsActive                     types.Bool   `tfsdk:"is_active"`
	AllowPullRequests            types.Bool   `tfsdk:"allow_pull_requests"`
	AllowDeployments             types.Bool   `tfsdk:"allow_deployments"`
	ConfigFile                   types.String `tfsdk:"config_file"`
	CancelPreviousPipelineEvents types.Set    `tfsdk:"cancel_previous_pipeline_events"`
	NetrcTrustedPlugins          types.Set    `tfsdk:"netrc_trusted_plugins"`
}

var repositoryModelTrustedAttributes = map[string]attr.Type{
	"network":  types.BoolType,
	"volumes":  types.BoolType,
	"security": types.BoolType,
}

func (m *repositoryModel) setValues(ctx context.Context, repo *woodpecker.Repo) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(repo.ID)
	m.ForgeID = types.Int64Value(repo.ForgeID)
	m.ForgeRemoteID = types.StringValue(repo.ForgeRemoteID)
	m.Owner = types.StringValue(repo.Owner)
	m.Name = types.StringValue(repo.Name)
	m.FullName = types.StringValue(repo.FullName)
	m.AvatarURL = types.StringValue(repo.Avatar)
	m.ForgeURL = types.StringValue(repo.ForgeURL)
	m.CloneURL = types.StringValue(repo.Clone)
	m.DefaultBranch = types.StringValue(repo.Branch)
	m.Timeout = types.Int64Value(repo.Timeout)
	m.Visibility = types.StringValue(repo.Visibility.String())
	m.IsPrivate = types.BoolValue(repo.IsSCMPrivate)
	m.Trusted, diags = types.ObjectValue(
		repositoryModelTrustedAttributes,
		map[string]attr.Value{
			"network":  types.BoolValue(repo.Trusted.Network),
			"volumes":  types.BoolValue(repo.Trusted.Volumes),
			"security": types.BoolValue(repo.Trusted.Security),
		},
	)
	diagsRes.Append(diags...)
	m.RequireApproval = types.StringValue(repo.RequireApproval.String())
	m.IsActive = types.BoolValue(repo.IsActive)
	m.AllowPullRequests = types.BoolValue(repo.AllowPullRequests)
	m.AllowDeployments = types.BoolValue(repo.AllowDeployments)
	m.ConfigFile = types.StringValue(repo.Config)
	m.CancelPreviousPipelineEvents, diags = types.SetValueFrom(ctx, types.StringType, repo.CancelPreviousPipelineEvents)
	diagsRes.Append(diags...)
	m.NetrcTrustedPlugins, diags = types.SetValueFrom(ctx, types.StringType, repo.NetrcTrustedPlugins)
	diagsRes.Append(diags...)

	return diagsRes
}

type trustedConfigurationPatchModel struct {
	Network  types.Bool `tfsdk:"network"`
	Volumes  types.Bool `tfsdk:"volumes"`
	Security types.Bool `tfsdk:"security"`
}

func (m *repositoryModel) toWoodpeckerPatch(ctx context.Context) (*woodpecker.RepoPatch, diag.Diagnostics) {
	var diags diag.Diagnostics

	repo := &woodpecker.RepoPatch{
		Config:            m.ConfigFile.ValueStringPointer(),
		Timeout:           m.Timeout.ValueInt64Pointer(),
		AllowPullRequests: m.AllowPullRequests.ValueBoolPointer(),
		AllowDeployments:  m.AllowDeployments.ValueBoolPointer(),
	}

	if visibility := m.Visibility.ValueStringPointer(); visibility != nil {
		converted := woodpecker.VisibilityMode(*visibility)
		repo.Visibility = &converted
	}

	diags.Append(m.CancelPreviousPipelineEvents.ElementsAs(ctx, &repo.CancelPreviousPipelineEvents, false)...)
	diags.Append(m.NetrcTrustedPlugins.ElementsAs(ctx, &repo.NetrcTrustedPlugins, true)...)

	var trusted *trustedConfigurationPatchModel
	diags.Append(m.Trusted.As(ctx, &trusted, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if trusted != nil {
		repo.Trusted = &woodpecker.TrustedConfigurationPatch{
			Network:  trusted.Network.ValueBoolPointer(),
			Volumes:  trusted.Volumes.ValueBoolPointer(),
			Security: trusted.Security.ValueBoolPointer(),
		}
	}

	return repo, diags
}

type repositorySecretResourceModelV0 struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Name         types.String `tfsdk:"name"`
	Value        types.String `tfsdk:"value"`
	Images       types.Set    `tfsdk:"images"`
	PluginsOnly  types.Bool   `tfsdk:"plugins_only"`
	Events       types.Set    `tfsdk:"events"`
}

type repositorySecretResourceModelV1 struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Name         types.String `tfsdk:"name"`
	Value        types.String `tfsdk:"value"`
	Images       types.Set    `tfsdk:"images"`
	Events       types.Set    `tfsdk:"events"`
}

func (m *repositorySecretResourceModelV1) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

func (m *repositorySecretResourceModelV1) toWoodpeckerModel(
	ctx context.Context,
) (*woodpecker.Secret, diag.Diagnostics) {
	var diags diag.Diagnostics

	secret := &woodpecker.Secret{
		ID:    m.ID.ValueInt64(),
		Name:  m.Name.ValueString(),
		Value: m.Value.ValueString(),
	}
	diags.Append(m.Images.ElementsAs(ctx, &secret.Images, false)...)
	diags.Append(m.Events.ElementsAs(ctx, &secret.Events, false)...)

	return secret, diags
}

type repositorySecretDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Name         types.String `tfsdk:"name"`
	Images       types.Set    `tfsdk:"images"`
	Events       types.Set    `tfsdk:"events"`
}

func (m *repositorySecretDataSourceModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

type repositoryCronModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	CreatorID    types.Int64  `tfsdk:"creator_id"`
	Schedule     types.String `tfsdk:"schedule"`
	CreatedAt    types.Int64  `tfsdk:"created_at"`
	Branch       types.String `tfsdk:"branch"`
}

func (m *repositoryCronModel) setValues(_ context.Context, cron *woodpecker.Cron) diag.Diagnostics {
	m.ID = types.Int64Value(cron.ID)
	m.Name = types.StringValue(cron.Name)
	m.RepositoryID = types.Int64Value(cron.RepoID)
	m.CreatorID = types.Int64Value(cron.CreatorID)
	m.Schedule = types.StringValue(cron.Schedule)
	m.CreatedAt = types.Int64Value(cron.Created)
	m.Branch = types.StringValue(cron.Branch)
	return nil
}

func (m *repositoryCronModel) toWoodpeckerModel(_ context.Context) (*woodpecker.Cron, diag.Diagnostics) {
	return &woodpecker.Cron{
		ID:        m.ID.ValueInt64(),
		Name:      m.Name.ValueString(),
		RepoID:    m.RepositoryID.ValueInt64(),
		CreatorID: m.CreatorID.ValueInt64(),
		Schedule:  m.Schedule.ValueString(),
		Created:   m.CreatedAt.ValueInt64(),
		Branch:    m.Branch.ValueString(),
	}, nil
}

type repositoryRegistryResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Address      types.String `tfsdk:"address"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (m *repositoryRegistryResourceModel) setValues(_ context.Context, registry *woodpecker.Registry) diag.Diagnostics {
	m.ID = types.Int64Value(registry.ID)
	m.Address = types.StringValue(registry.Address)
	m.Username = types.StringValue(registry.Username)
	return nil
}

func (m *repositoryRegistryResourceModel) toWoodpeckerModel(
	_ context.Context,
) (*woodpecker.Registry, diag.Diagnostics) {
	return &woodpecker.Registry{
		ID:       m.ID.ValueInt64(),
		Address:  m.Address.ValueString(),
		Username: m.Username.ValueString(),
		Password: m.Password.ValueString(),
	}, nil
}

type repositoryRegistryDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Address      types.String `tfsdk:"address"`
	Username     types.String `tfsdk:"username"`
}

func (m *repositoryRegistryDataSourceModel) setValues(
	_ context.Context,
	registry *woodpecker.Registry,
) diag.Diagnostics {
	m.ID = types.Int64Value(registry.ID)
	m.Address = types.StringValue(registry.Address)
	m.Username = types.StringValue(registry.Username)
	return nil
}
