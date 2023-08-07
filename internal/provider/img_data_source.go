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
	"fmt"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/congop/terraform-provider-saya/internal/slices"
	saya_types "github.com/congop/terraform-provider-saya/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ImageDataSource{}

func NewExampleDataSource() datasource.DataSource {
	return &ImageDataSource{}
}

// ImageDataSource defines the data source implementation.
type ImageDataSource struct {
	sayaExeCtx *SayaExecutionCtx
}

// ImageDataSourceModel describes the data source data model.
type ImageDataSourceModel struct {
	Name      types.String `tfsdk:"name"`
	ImgType   types.String `tfsdk:"img_type"`
	Platform  types.String `tfsdk:"platform"`
	Id        types.String `tfsdk:"id"`
	Sha256    types.String `tfsdk:"sha256"`
	OsVariant types.String `tfsdk:"os_variant"`
	Filters   types.Map    `tfsdk:"filters"`
}

func (d *ImageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (d *ImageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source for an image in the local image store",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "image identifier",
			},
			"sha256": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "image sha256",
			},
			"os_variant": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "image os-variant e.g. ubuntu, alpine, debian, etc.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "image name[:tag], e.r. appserver:v1",
				Description:         "image name[:tag], e.g. appserver:v1",
				Required:            true,
				Validators:          []validator.String{saya_types.SayaRefMustBeNameVersion{}},
				// PlanModifiers:       []planmodifier.String{saya_types.SayaRefNormalizeModifier{}},
				// CustomType:          saya_types.ReferenceTfType{},
				// Computed:            true,
			},
			"img_type": schema.StringAttribute{
				MarkdownDescription: "image type, e.g. ova|vmdk, vhd, img, qcow2, etc.",
				Required:            true,
				// Default:             stringdefault.StaticString("example value when not configured"),
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "image platform, e.g linux/amd64, linux/arm64/v7, defaults to host platform",
				Description:         "image platform, e.g linux/amd64, linux/arm64/v7, defaults to host platform",
				Optional:            true,
				// TODO Default:             defaults.String(""),
			},

			"filters": schema.MapAttribute{
				ElementType:         types.SetType{ElemType: types.StringType},
				MarkdownDescription: "filter criteria to match the image with, e.g label='audience=tester' ",
				Description:         "image platform, e.g linux/amd64, linux/arm64/v7, defaults to host platform",
				Optional:            true,
			},
		},
	}
}

func (d *ImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sayaExeCtx, ok := req.ProviderData.(SayaExecutionCtx)
	if !ok {
		resp.Diagnostics.AddError(
			"ImageDataSource.Configure -- unexpected provider data type",
			fmt.Sprintf(
				"ImageDataSource.Configure -- unexpected provider data type: expected=%T got=%T",
				SayaExecutionCtx{}, req.ProviderData))
		return
	}
	d.sayaExeCtx = &sayaExeCtx
}

func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// find image in the for according to the request setting and set properties from image info
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lsReq := saya.LsRequest{
		Name:      data.Name.ValueString(),
		ImgType:   data.ImgType.ValueString(),
		Platform:  data.Platform.ValueString(),
		OsVariant: data.OsVariant.ValueString(),
		Sha256:    data.Sha256.ValueString(),

		RequestSayaCtx: d.sayaExeCtx.ToRequestSayaCtx(),
	}

	switch lsResList, err := saya.Ls(ctx, lsReq); {
	case err != nil:
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	case len(lsResList) == 1:
		lsRes := lsResList[0]
		// save into the Terraform state.
		data.Id = types.StringValue(lsRes.Sha256)
		data.ImgType = types.StringValue(lsRes.Type)
		data.Name = types.StringValue(lsRes.Name + ":" + lsRes.Version)
		data.Platform = types.StringValue(lsRes.Platform.PlatformStr())
		data.Sha256 = types.StringValue(lsRes.Sha256)
	case len(lsResList) == 0:
		resp.Diagnostics.AddError(
			"No image found in the forge",
			fmt.Sprintf("No image found in the forge(local image store): request-data=%#v", lsReq))
		return
	default:
		lsResListStr := slices.MapMust(lsResList, saya.LsResult.PlatformNameVersionTypeTaglike)
		resp.Diagnostics.AddError("Too many images found", fmt.Sprintf("Too many images found: found=%v", lsResListStr))
		return
	}

	log.Tracef(ctx, "ImageDataSource.Configure -- read a data source: ls-result=%#v", lsReq)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
