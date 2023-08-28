package internal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

type userModel struct {
	ID     types.Int64  `tfsdk:"id"`
	Login  types.String `tfsdk:"login"`
	Email  types.String `tfsdk:"email"`
	Avatar types.String `tfsdk:"avatar"`
	Active types.Bool   `tfsdk:"active"`
	Admin  types.Bool   `tfsdk:"admin"`
}

func (m *userModel) setValues(_ context.Context, user *woodpecker.User) diag.Diagnostics {
	m.ID = types.Int64Value(user.ID)
	m.Login = types.StringValue(user.Login)
	m.Email = types.StringValue(user.Email)
	m.Avatar = types.StringValue(user.Avatar)
	m.Active = types.BoolValue(user.Active)
	m.Admin = types.BoolValue(user.Admin)
	return nil
}

func (m *userModel) toWoodpeckerModel(_ context.Context) (*woodpecker.User, diag.Diagnostics) {
	return &woodpecker.User{
		ID:     m.ID.ValueInt64(),
		Login:  m.Login.ValueString(),
		Email:  m.Email.ValueString(),
		Avatar: m.Avatar.ValueString(),
		Active: m.Active.ValueBool(),
		Admin:  m.Admin.ValueBool(),
	}, nil
}

type secretModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Images      types.Set    `tfsdk:"images"`
	PluginsOnly types.Bool   `tfsdk:"plugins_only"`
	Events      types.Set    `tfsdk:"events"`
}

func (m *secretModel) setValues(ctx context.Context, secret *woodpecker.Secret) diag.Diagnostics {
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

func (m *secretModel) toWoodpeckerModel(ctx context.Context) (*woodpecker.Secret, diag.Diagnostics) {
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
