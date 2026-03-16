package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pleme-io/terraform-provider-akeyless-gen/internal/resources"
)

var _ provider.Provider = &AkeylessProvider{}

// AkeylessProvider defines the provider implementation.
type AkeylessProvider struct {
	version string
}

// AkeylessProviderModel describes the provider data model.
type AkeylessProviderModel struct {
	ApiGatewayAddress types.String `tfsdk:"api_gateway_address"`
	AccessToken       types.String `tfsdk:"access_token"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AkeylessProvider{
			version: version,
		}
	}
}

func (p *AkeylessProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "akeyless"
	resp.Version = p.version
}

func (p *AkeylessProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Akeyless Vault Provider — manage secrets, auth methods, targets, and roles",
		Attributes: map[string]schema.Attribute{
			"api_gateway_address": schema.StringAttribute{
				Description: "Akeyless API gateway URL. Can also be set via AKEYLESS_GATEWAY env var.",
				Optional:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "Akeyless access token. Can also be set via AKEYLESS_ACCESS_TOKEN env var.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *AkeylessProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AkeylessProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	gatewayURL := config.ApiGatewayAddress.ValueString()
	if gatewayURL == "" {
		gatewayURL = os.Getenv("AKEYLESS_GATEWAY")
	}
	if gatewayURL == "" {
		gatewayURL = "https://api.akeyless.io"
	}

	token := config.AccessToken.ValueString()
	if token == "" {
		token = os.Getenv("AKEYLESS_ACCESS_TOKEN")
	}

	client := resources.NewAkeylessClient(gatewayURL, token)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AkeylessProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewStaticSecretResource,
	}
}

func (p *AkeylessProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
