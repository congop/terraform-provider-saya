// Copyright (C) 2023 Patrice Congo <@congop>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"strings"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/opaque"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &SayaProvider{}

// SayaProvider defines the provider implementation.
type SayaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	logLevel string
}

type SayaExecutionCtx struct {
	SayaExe    string        // saya executable command or path, default to saya
	Config     string        // saya yaml config path
	Forge      string        // forge(local image store+ work directory, etc.) path
	LicenseKey opaque.String // License key
	LogLevel   string

	repos *saya.Repos
}

func (exeCtx *SayaExecutionCtx) setHttpRepo(httpRepo *saya.HttpRepo) {
	if httpRepo == nil {
		return
	}
	if exeCtx.repos == nil {
		exeCtx.repos = &saya.Repos{}
	}
	exeCtx.repos.Http = httpRepo
}

func (exeCtx *SayaExecutionCtx) HttpRepo() *saya.HttpRepo {
	if exeCtx.repos == nil {
		return nil
	}
	return exeCtx.repos.Http
}

// SayaProviderModel describes the provider data model.
// @mind no LogLevel because terraform sets it throw environment variable so we are following the lead.
type SayaProviderModel struct {
	Exe        types.String `tfsdk:"exe"`
	Config     types.String `tfsdk:"config"`
	Forge      types.String `tfsdk:"forge"`
	LicenseKey types.String `tfsdk:"license_key"`

	HttpRepo types.Object `tfsdk:"http_repo"`
}

func (p *SayaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "saya"
	resp.Version = p.version
}

func (p *SayaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config": schema.StringAttribute{
				MarkdownDescription: "saya yaml config path",
				Optional:            true,
			},
			"exe": schema.StringAttribute{
				MarkdownDescription: "saya exe command or path; defaults to saya",
				Optional:            true,
			},
			"forge": schema.StringAttribute{
				MarkdownDescription: "forge location",
				Optional:            true,
			},
			"license_key": schema.StringAttribute{
				MarkdownDescription: "license key value or file",
				Optional:            true,
				Sensitive:           true,
			},
			"http_repo": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "the url  of the remote repository",
						Required:    true,
					},
					"base_path": schema.StringAttribute{
						Description: "the base path  of the remote repository",
						Optional:    true,
					},
					"upload_strategy": schema.StringAttribute{
						Description: "upload strategy",
						Optional:    true,
					},
					"basic_auth": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Description: "the username to authenticate with",
								Required:    true,
							},
							"password": schema.StringAttribute{
								Description: "the password",
								Required:    true,
							},
						},
						Optional: true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (p *SayaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SayaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	exeCtx := SayaExecutionCtx{
		SayaExe:    strings.TrimSpace(data.Exe.ValueString()),
		Config:     strings.TrimSpace(data.Config.ValueString()),
		Forge:      strings.TrimSpace(data.Forge.ValueString()),
		LicenseKey: *opaque.NewString(strings.TrimSpace(data.LicenseKey.ValueString())),
		LogLevel:   p.logLevel,
	}

	if exeCtx.SayaExe == "" {
		exeCtx.SayaExe = "saya"
	}

	httpRepoTf := &ImageResourceModelHttpRepo{}

	// opts to avoid <<Received null value, however the target type cannot handle null values.>>
	if !(data.HttpRepo.IsNull() || data.HttpRepo.IsUnknown()) {
		diagsMapping := data.HttpRepo.As(ctx, httpRepoTf, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		if diagsMapping.HasError() {
			resp.Diagnostics.Append(diagsMapping...)
			return
		}
		httpRepoSaya := saya.HttpRepo{
			RepoUrl:        httpRepoTf.RepoUrl,
			BasePath:       httpRepoTf.BasePath,
			UploadStrategy: httpRepoTf.UploadStrategy,
			AuthHttpBasic: saya.AuthHttpBasic{
				Username: httpRepoTf.AuthHttpBasic.Username,
				Pwd:      httpRepoTf.AuthHttpBasic.Pwd,
			},
		}

		exeCtx.setHttpRepo(httpRepoSaya.NormalizeToNil())

	}

	resp.DataSourceData = exeCtx
	resp.ResourceData = exeCtx
}

func (p *SayaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewImageResource,
	}
}

func (p *SayaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func New(version string, logLevel string) func() provider.Provider {
	ctx := tfsdklog.NewRootProviderLogger(context.Background(), tfsdklog.WithLevel(hclog.LevelFromString(logLevel)))
	log.Debugf(ctx, "New -- provider requested:  version=%s, loglevel=%s:",
		version, logLevel)
	return func() provider.Provider {
		return &SayaProvider{
			version:  version,
			logLevel: logLevel,
		}
	}
}
