package internal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type userModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Login     types.String `tfsdk:"login"`
	Email     types.String `tfsdk:"email"`
	AvatarURL types.String `tfsdk:"avatar_url"`
	Active    types.Bool   `tfsdk:"is_active"`
	Admin     types.Bool   `tfsdk:"is_admin"`
}

func (m *userModel) setValues(_ context.Context, user *woodpecker.User) diag.Diagnostics {
	m.ID = types.Int64Value(user.ID)
	m.Login = types.StringValue(user.Login)
	m.Email = types.StringValue(user.Email)
	m.AvatarURL = types.StringValue(user.Avatar)
	m.Active = types.BoolValue(user.Active)
	m.Admin = types.BoolValue(user.Admin)
	return nil
}

func (m *userModel) toWoodpeckerModel(_ context.Context) (*woodpecker.User, diag.Diagnostics) {
	return &woodpecker.User{
		ID:     m.ID.ValueInt64(),
		Login:  m.Login.ValueString(),
		Email:  m.Email.ValueString(),
		Avatar: m.AvatarURL.ValueString(),
		Active: m.Active.ValueBool(),
		Admin:  m.Admin.ValueBool(),
	}, nil
}

type secretResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Images      types.Set    `tfsdk:"images"`
	PluginsOnly types.Bool   `tfsdk:"plugins_only"`
	Events      types.Set    `tfsdk:"events"`
}

func (m *secretResourceModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.PluginsOnly = types.BoolValue(secret.PluginsOnly)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

func (m *secretResourceModel) toWoodpeckerModel(ctx context.Context) (*woodpecker.Secret, diag.Diagnostics) {
	var diags diag.Diagnostics

	secret := &woodpecker.Secret{
		ID:          m.ID.ValueInt64(),
		Name:        m.Name.ValueString(),
		Value:       m.Value.ValueString(),
		PluginsOnly: m.PluginsOnly.ValueBool(),
	}
	diags.Append(m.Images.ElementsAs(ctx, &secret.Images, false)...)
	diags.Append(m.Events.ElementsAs(ctx, &secret.Events, false)...)

	return secret, diags
}

type secretDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Images      types.Set    `tfsdk:"images"`
	PluginsOnly types.Bool   `tfsdk:"plugins_only"`
	Events      types.Set    `tfsdk:"events"`
}

func (m *secretDataSourceModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.PluginsOnly = types.BoolValue(secret.PluginsOnly)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

type repositoryModel struct {
	ID                types.Int64  `tfsdk:"id"`
	ForgeRemoteID     types.String `tfsdk:"forge_remote_id"`
	Owner             types.String `tfsdk:"owner"`
	Name              types.String `tfsdk:"name"`
	FullName          types.String `tfsdk:"full_name"`
	AvatarURL         types.String `tfsdk:"avatar_url"`
	URL               types.String `tfsdk:"url"`
	CloneURL          types.String `tfsdk:"clone_url"`
	DefaultBranch     types.String `tfsdk:"default_branch"`
	SCMKind           types.String `tfsdk:"scm"`
	Timeout           types.Int64  `tfsdk:"timeout"`
	Visibility        types.String `tfsdk:"visibility"`
	IsSCMPrivate      types.Bool   `tfsdk:"is_private"`
	IsTrusted         types.Bool   `tfsdk:"is_trusted"`
	IsGated           types.Bool   `tfsdk:"is_gated"`
	AllowPullRequests types.Bool   `tfsdk:"allow_pull_requests"`
	ConfigFile        types.String `tfsdk:"config_file"`
	NetrcOnlyTrusted  types.Bool   `tfsdk:"netrc_only_trusted"`
}

func (m *repositoryModel) setValues(_ context.Context, repo *woodpecker.Repo) diag.Diagnostics {
	m.ID = types.Int64Value(repo.ID)
	m.ForgeRemoteID = types.StringValue(repo.ForgeRemoteID)
	m.Owner = types.StringValue(repo.Owner)
	m.Name = types.StringValue(repo.Name)
	m.FullName = types.StringValue(repo.FullName)
	m.AvatarURL = types.StringValue(repo.Avatar)
	m.URL = types.StringValue(repo.Link)
	m.CloneURL = types.StringValue(repo.Clone)
	m.DefaultBranch = types.StringValue(repo.DefaultBranch)
	m.SCMKind = types.StringValue(repo.SCMKind)
	m.Timeout = types.Int64Value(repo.Timeout)
	m.Visibility = types.StringValue(repo.Visibility)
	m.IsSCMPrivate = types.BoolValue(repo.IsSCMPrivate)
	m.IsTrusted = types.BoolValue(repo.IsTrusted)
	m.IsGated = types.BoolValue(repo.IsGated)
	m.AllowPullRequests = types.BoolValue(repo.AllowPullRequests)
	m.ConfigFile = types.StringValue(repo.Config)
	m.NetrcOnlyTrusted = types.BoolValue(repo.NetrcOnlyTrusted)
	return nil
}

func (m *repositoryModel) toWoodpeckerPatch(_ context.Context) (*woodpecker.RepoPatch, diag.Diagnostics) {
	return &woodpecker.RepoPatch{
		Config:     m.ConfigFile.ValueStringPointer(),
		IsTrusted:  m.IsTrusted.ValueBoolPointer(),
		IsGated:    m.IsGated.ValueBoolPointer(),
		Timeout:    m.Timeout.ValueInt64Pointer(),
		Visibility: m.Visibility.ValueStringPointer(),
		AllowPull:  m.AllowPullRequests.ValueBoolPointer(),
	}, nil
}

type repositorySecretResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Name         types.String `tfsdk:"name"`
	Value        types.String `tfsdk:"value"`
	Images       types.Set    `tfsdk:"images"`
	PluginsOnly  types.Bool   `tfsdk:"plugins_only"`
	Events       types.Set    `tfsdk:"events"`
}

func (m *repositorySecretResourceModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.PluginsOnly = types.BoolValue(secret.PluginsOnly)
	m.Events, diags = types.SetValueFrom(ctx, types.StringType, secret.Events)
	diagsRes.Append(diags...)

	return diagsRes
}

func (m *repositorySecretResourceModel) toWoodpeckerModel(ctx context.Context) (*woodpecker.Secret, diag.Diagnostics) {
	var diags diag.Diagnostics

	secret := &woodpecker.Secret{
		ID:          m.ID.ValueInt64(),
		Name:        m.Name.ValueString(),
		Value:       m.Value.ValueString(),
		PluginsOnly: m.PluginsOnly.ValueBool(),
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
	PluginsOnly  types.Bool   `tfsdk:"plugins_only"`
	Events       types.Set    `tfsdk:"events"`
}

func (m *repositorySecretDataSourceModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
	var diagsRes diag.Diagnostics
	var diags diag.Diagnostics

	m.ID = types.Int64Value(secret.ID)
	m.Name = types.StringValue(secret.Name)
	m.Images, diags = types.SetValueFrom(ctx, types.StringType, secret.Images)
	diagsRes.Append(diags...)
	m.PluginsOnly = types.BoolValue(secret.PluginsOnly)
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
	Email        types.String `tfsdk:"email"`
}

func (m *repositoryRegistryResourceModel) setValues(_ context.Context, registry *woodpecker.Registry) diag.Diagnostics {
	m.ID = types.Int64Value(registry.ID)
	m.Address = types.StringValue(registry.Address)
	m.Username = types.StringValue(registry.Username)
	m.Email = types.StringValue(registry.Email)
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
		Email:    m.Email.ValueString(),
	}, nil
}

type repositoryRegistryDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
	Address      types.String `tfsdk:"address"`
	Username     types.String `tfsdk:"username"`
	Email        types.String `tfsdk:"email"`
}

func (m *repositoryRegistryDataSourceModel) setValues(
	_ context.Context,
	registry *woodpecker.Registry,
) diag.Diagnostics {
	m.ID = types.Int64Value(registry.ID)
	m.Address = types.StringValue(registry.Address)
	m.Username = types.StringValue(registry.Username)
	m.Email = types.StringValue(registry.Email)
	return nil
}
