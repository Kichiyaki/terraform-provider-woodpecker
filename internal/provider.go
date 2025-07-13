package internal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/oauth2"
)

const importStateIDSeparator = "/"

type woodpeckerProvider struct {
	version string
}

var _ provider.Provider = (*woodpeckerProvider)(nil)

func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &woodpeckerProvider{
			version: version,
		}
	}
}

func (p *woodpeckerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "woodpecker"
	resp.Version = p.version
}

func (p *woodpeckerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Terraform provider used to interact with Woodpecker CI resources.",
		MarkdownDescription: "A Terraform provider used to interact with" +
			" [Woodpecker CI](https://woodpecker-ci.org/) resources." +
			"\n\n\n- v0.4.x and later versions of the provider work with Woodpecker 3.x+" +
			"\n- v0.2.x and v0.3.x versions of the provider work with Woodpecker 3.0.0>1.x>=2.0.0" +
			"\n- v0.1.x version of the provider works with Woodpecker 2.0.0>1.x>=1.0.0",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Optional: true,
				Description: `This is the target Woodpecker CI base API endpoint. It must be provided, but
					can also be sourced from the WOODPECKER_SERVER environment
					variable.`,
			},
			"token": schema.StringAttribute{
				Optional: true,
				Description: `A Woodpecker CI Personal Access Token. It must be provided, but
					can also be sourced from the WOODPECKER_TOKEN environment
					variable.`,
			},
		},
	}
}

func (p *woodpeckerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newUserDataSource,
		newSecretDataSource,
		newOrgDataSource,
		newOrgSecretDataSource,
		newRepositoryDataSource,
		newRepositorySecretDataSource,
		newRepositoryCronDataSource,
		newRepositoryRegistryDataSource,
	}
}

func (p *woodpeckerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newUserResource,
		newSecretResource,
		newOrgSecretResource,
		newRepositoryResource,
		newRepositorySecretResource,
		newRepositoryCronResource,
		newRepositoryRegistryResource,
	}
}

func (p *woodpeckerProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	cfg := newProviderConfig(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	client := newClient(ctx, cfg, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

type providerConfig struct {
	Server types.String `tfsdk:"server"`
	Token  types.String `tfsdk:"token"`
}

func newProviderConfig(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) providerConfig {
	var config providerConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return config
	}

	if config.Server.ValueString() == "" {
		config.Server = types.StringValue(os.Getenv("WOODPECKER_SERVER"))
	}

	if config.Server.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Server URL Configuration",
			"While configuring the provider, the server URL was not found in "+
				"the WOODPECKER_SERVER environment variable or provider "+
				"configuration block server attribute.",
		)
	}

	if config.Token.ValueString() == "" {
		config.Token = types.StringValue(os.Getenv("WOODPECKER_TOKEN"))
	}

	if config.Token.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the WOODPECKER_TOKEN environment variable or provider "+
				"configuration block token attribute.",
		)
	}

	return config
}

func newClient(
	ctx context.Context,
	config providerConfig,
	resp *provider.ConfigureResponse,
) woodpecker.Client {
	client := woodpecker.NewClient(
		config.Server.ValueString(),
		(&oauth2.Config{}).Client(ctx, &oauth2.Token{
			AccessToken: config.Token.ValueString(),
		}),
	)

	_, err := client.Self()
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get current user", err.Error())
		return nil
	}

	c, err := semver.NewConstraint(">= 3.0.0")
	if err != nil {
		resp.Diagnostics.AddError(
			"Couldn't parse woodpecker version constraint",
			fmt.Sprintf(
				"%s. Please report this issue to the provider developers.",
				err,
			),
		)
		return nil
	}

	ver, err := client.Version()
	if err != nil {
		resp.Diagnostics.AddError("Couldn't get woodpecker version", err.Error())
		return nil
	}

	// split is required because in some cases the version looks like this: 2.0.0-f05c1631d2
	parsedVer, err := semver.NewVersion(strings.Split(ver.Version, "-")[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Couldn't parse woodpecker version",
			fmt.Sprintf(
				"%s. Please report this issue to the provider developers.",
				err,
			),
		)
		return nil
	}

	if !c.Check(parsedVer) {
		resp.Diagnostics.AddError(
			"Woodpecker version doesn't satisfy the constraint",
			fmt.Sprintf(
				"Current woodpecker version: %s, expected: %s."+
					" Consider using an older version of the provider or update your Woodpecker CI instance.",
				ver.Version,
				c.String(),
			),
		)
		return nil
	}

	return client
}
