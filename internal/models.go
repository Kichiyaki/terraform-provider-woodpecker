package internal

import (
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

func (m *userModel) setValues(user *woodpecker.User) {
	m.ID = types.Int64Value(user.ID)
	m.Login = types.StringValue(user.Login)
	m.Email = types.StringValue(user.Email)
	m.Avatar = types.StringValue(user.Avatar)
	m.Active = types.BoolValue(user.Active)
	m.Admin = types.BoolValue(user.Admin)
}

func (m *userModel) toWoodpeckerModel() *woodpecker.User {
	return &woodpecker.User{
		ID:     m.ID.ValueInt64(),
		Login:  m.Login.ValueString(),
		Email:  m.Email.ValueString(),
		Avatar: m.Avatar.ValueString(),
		Active: m.Active.ValueBool(),
		Admin:  m.Admin.ValueBool(),
	}
}
