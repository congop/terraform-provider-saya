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
	"strings"

	"github.com/aws/smithy-go/time"
	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/congop/terraform-provider-saya/internal/slices"
	saya_types "github.com/congop/terraform-provider-saya/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pkg/errors"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ImageResource{}
var _ resource.ResourceWithImportState = &ImageResource{}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

// ImageResource defines the resource implementation.
type ImageResource struct {
	sayaExeCtx *SayaExecutionCtx
}

// ImageResourceModel describes the resource data model.
type ImageResourceModel struct {
	Name        types.String `tfsdk:"name"`
	ImgType     types.String `tfsdk:"img_type"`
	Platform    types.String `tfsdk:"platform"`
	Hash        types.String `tfsdk:"hash"`
	Id          types.String `tfsdk:"id"`
	Sha256      types.String `tfsdk:"sha256"`
	KeepLocally types.Bool   `tfsdk:"keep_locally"`
	RepoType    types.String `tfsdk:"repo_type"`
	HttpRepo    types.Object `tfsdk:"http_repo"`
}
type ImageResourceModelHttpAuthBasic struct {
	Username string `tfsdk:"username"`
	Pwd      string `tfsdk:"password"`
}

type ImageResourceModelHttpRepo struct {
	RepoUrl        string                          `tfsdk:"url"`
	BasePath       string                          `tfsdk:"base_path"`
	UploadStrategy string                          `tfsdk:"upload_strategy"`
	AuthHttpBasic  ImageResourceModelHttpAuthBasic `tfsdk:"basic_auth"`
}

type ImageResourceModelS3RepoCred struct {
	AccessKeyID     string `tfsdk:"access_key_id"`
	SecretAccessKey string `tfsdk:"secret_access_key"`
	SessionToken    string `tfsdk:"session_token"`
	Source          string `tfsdk:"source"`
	CanExpire       bool   `tfsdk:"can_expire"`
	Expires         string `tfsdk:"expires"`
}

func (credTf *ImageResourceModelS3RepoCred) AsSayaCred() (*saya.AwsCredentials, error) {
	if credTf == nil {
		return nil, nil
	}
	sayaCred := saya.AwsCredentials{
		AccessKeyID:     credTf.AccessKeyID,
		SecretAccessKey: credTf.SecretAccessKey,
		SessionToken:    credTf.SessionToken,
		Source:          credTf.Source,
		CanExpire:       credTf.CanExpire,
		Expires:         nil,
	}

	if expiresTf := strings.TrimSpace(credTf.Expires); expiresTf != "" {
		expires, err := time.ParseDateTime(expiresTf)
		if err != nil {
			return nil, errors.Wrapf(err,
				"ImageResourceModelS3RepoCred.AsSayaCred -- bad date time format for expires"+
					"\n\texpected-format: RFC3339, e.g. 1985-04-12T23:20:50.52Z"+
					"\n\tvalue-string=%s \n\tparse-issue=%s",
				expiresTf, err.Error())
		}
		sayaCred.Expires = &expires
	}

	return &sayaCred, nil
}

type ImageResourceModelS3Repo struct {
	Bucket       string                        `tfsdk:"bucket"`
	BaseKey      string                        `tfsdk:"base_key"`
	EpUrlS3      string                        `tfsdk:"ep_url_s3"`
	Region       string                        `tfsdk:"region"`
	UsePathStyle bool                          `tfsdk:"use_path_style"`
	Credentials  *ImageResourceModelS3RepoCred `tfsdk:"credentials"`
}

func (repo ImageResourceModelHttpRepo) NormalizeToNil() *ImageResourceModelHttpRepo {
	repo.AuthHttpBasic.Pwd = strings.TrimSpace(repo.AuthHttpBasic.Pwd)
	repo.AuthHttpBasic.Username = strings.TrimSpace(repo.AuthHttpBasic.Username)
	repo.BasePath = strings.TrimSpace(repo.BasePath)
	repo.RepoUrl = strings.TrimSpace(repo.RepoUrl)
	repo.UploadStrategy = strings.TrimSpace(repo.UploadStrategy)

	if (repo == ImageResourceModelHttpRepo{}) {
		return nil
	}
	return &repo
}

func (r *ImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	log.Debugf(ctx, "ImageResource.MetaData=%#v\n\n", req)
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *ImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pull a saya Image to the host from the configure repository",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "image identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sha256": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "image sha256",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"hash": schema.StringAttribute{
				MarkdownDescription: "hash to verify the pulled image, format <hash-type>:<hash-value>; md5:7287292, sha256:1234555",
				Description:         "hash to verify the pulled image, format <hash-type>:<hash-value>; md5:7287292, sha256:1234555",
				Optional:            true,
			},
			"keep_locally": schema.BoolAttribute{
				MarkdownDescription: "true to keep the pulled image in the local image store, false otherwise",
				Description:         "true to keep the pulled image in the local image store, false otherwise",
				Optional:            true,
			},
			"repo_type": schema.StringAttribute{
				MarkdownDescription: "the type of the remote repository; e.g.: http|s3",
				Description:         "the type of the remote repository; e.g.: http|s3",
				Optional:            true,
				// PlanModifiers: []planmodifier.String{
				// 	SuppressFromDiffAttributePlanModifierStr("repo_type"),
				// },
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
				// PlanModifiers: []planmodifier.Object{
				// 	SyncAttributePlanModifier("http_repo"),
				// },
			},
		},
	}

}
func attributesHttpRepo() map[string]attr.Type {
	return map[string]attr.Type{
		"url":             basetypes.StringType{},
		"base_path":       basetypes.StringType{},
		"upload_strategy": basetypes.StringType{},
		"basic_auth": basetypes.ObjectType{
			AttrTypes: map[string]attr.Type{
				"username": basetypes.StringType{},
				"password": basetypes.StringType{},
			},
		},
		//  schema.SingleNestedAttribute{
		// 	Attributes: map[string]schema.Attribute{
		// 		"username": schema.StringAttribute{
		// 			Description: "the username to authenticate with",
		// 			Required:    true,
		// 		},
		// 		"password": schema.StringAttribute{
		// 			Description: "the password",
		// 			Required:    true,
		// 		},
		// 	},
		// 	Optional: true,
		// },
	}
}

func (r *ImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	// no provider data available or expected yet
	// may set provider saya setting (setting yaml)
	sayaExeCtx, ok := req.ProviderData.(SayaExecutionCtx)
	if !ok {
		resp.Diagnostics.AddError(
			"unexpected provider data type",
			fmt.Sprintf(
				"unexpected provider data type: expected=%T got=%T",
				SayaExecutionCtx{}, req.ProviderData))
		return
	}
	r.sayaExeCtx = &sayaExeCtx
}

func (r *ImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullReq := saya.PullRequest{
		Name:     data.Name.ValueString(),
		ImgType:  data.ImgType.ValueString(),
		Platform: data.Platform.ValueString(),
		Hash:     data.Hash.ValueString(),
		RepoType: data.RepoType.ValueString(),

		Exe:        r.sayaExeCtx.SayaExe,
		Config:     r.sayaExeCtx.Config,
		Forge:      r.sayaExeCtx.Forge,
		LicenseKey: r.sayaExeCtx.LicenseKey,
		LogLevel:   r.sayaExeCtx.LogLevel,
	}

	switch repoType := data.RepoType.ValueString(); {
	case saya.IsRepoTypeHttp(repoType):
		pullReq.HttpRepo = r.sayaExeCtx.HttpRepo()
	case saya.IsRepoTypeS3(repoType):
		pullReq.S3Repo = r.sayaExeCtx.S3Repo()
	case repoType == "" && r.sayaExeCtx.repos.HttpOnly():
		pullReq.HttpRepo = r.sayaExeCtx.HttpRepo()
	case repoType == "" && r.sayaExeCtx.repos.S3Only():
		pullReq.S3Repo = r.sayaExeCtx.S3Repo()
	case repoType == "":
		resp.Diagnostics.AddError(
			"repo-type not specified but not exactly one repo-type available",
			fmt.Sprintf(
				"repo-type not specified but not exactly one repo-type available: available=%s",
				r.sayaExeCtx.repos.AvailableRepoTypes()))
	default:
		resp.Diagnostics.AddError(
			"illegal saya repository configuration",
			fmt.Sprintf(
				"illegal saya repository configuration: \n\twanted-repo-type=%s \n\tavailable=%s",
				repoType, r.sayaExeCtx.repos.AvailableRepoTypes()))

	}

	pullRes, err := saya.Pull(ctx, pullReq)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}
	platformStr := pullRes.Platform.PlatformStr()
	id, err := saya.ImgId(pullRes.Name, pullRes.Version, platformStr, pullRes.Type)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}

	data.Id = types.StringValue(id)
	data.Sha256 = types.StringValue(pullRes.Sha256)
	data.ImgType = types.StringValue(pullRes.Type)
	data.Platform = types.StringValue(platformStr)
	// data.RepoType = types.StringValue(pullRes.SrcType)

	log.Tracef(ctx, "saya image tf-created: pullRes=%#v", pullRes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func pullImgAndUpdateData(
	ctx context.Context, r *ImageResource,
	data *ImageResourceModel, diags *diag.Diagnostics,
) {
	httpRepoSaya := r.sayaExeCtx.HttpRepo()

	pullReq := saya.PullRequest{
		Name:     data.Name.ValueString(),
		ImgType:  data.ImgType.ValueString(),
		Platform: data.Platform.ValueString(),
		Hash:     data.Hash.ValueString(),
		RepoType: data.RepoType.ValueString(),
		HttpRepo: httpRepoSaya.NormalizeToNil(),

		Exe:        r.sayaExeCtx.SayaExe,
		Config:     r.sayaExeCtx.Config,
		Forge:      r.sayaExeCtx.Forge,
		LicenseKey: r.sayaExeCtx.LicenseKey,
		LogLevel:   r.sayaExeCtx.LogLevel,
	}

	pullRes, err := saya.Pull(ctx, pullReq)
	if err != nil {
		diags.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}
	platformStr := pullRes.Platform.PlatformStr()
	id, err := saya.ImgId(pullRes.Name, pullRes.Version, platformStr, pullRes.Type)
	if err != nil {
		diags.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}

	data.Id = types.StringValue(id)
	data.Sha256 = types.StringValue(pullRes.Sha256)
	data.ImgType = types.StringValue(pullRes.Type)
	data.Platform = types.StringValue(platformStr)
	data.Name = types.StringValue(pullRes.Name + ":" + pullRes.Version)

}

func (r *ImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lsReq := saya.LsRequest{
		Name:      data.Name.ValueString(),
		ImgType:   data.ImgType.ValueString(),
		Platform:  data.Platform.ValueString(),
		OsVariant: "",
		Sha256:    "",

		RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx(),
	}
	idStr := data.Id.ValueString()
	importing := lsReq.Name == ""
	if importing {
		id, err := saya.ParseImgId(idStr)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
			return
		}
		lsReq.Name = id.R.Normalized()
		lsReq.ImgType = id.ImgType
		lsReq.Platform = id.P.PlatformStr()
	}

	switch lsResList, err := saya.Ls(ctx, lsReq); {
	case err != nil:
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	case len(lsResList) == 1:
		lsRes := lsResList[0]
		platformStr := lsRes.Platform.PlatformStr()
		id, err := saya.ImgId(lsRes.Name, lsRes.Version, platformStr, lsRes.Type)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
			return
		}

		if data.Id.IsNull() || data.Id.IsUnknown() {
			data.Id = types.StringValue(id)
		}
		data.Sha256 = types.StringValue(lsRes.Sha256)
		data.Name = types.StringValue(lsRes.Name + ":" + lsRes.Version)
		data.ImgType = types.StringValue(lsRes.Type)
		data.Platform = types.StringValue(platformStr)
		if importing {
			data.RepoType = types.StringValue(lsRes.SrcType)
		}
	case len(lsResList) == 0:
		resp.Diagnostics.AddError(
			"ImageResource -- No image found in the forge",
			fmt.Sprintf(
				"ImageResource -- No image found in the forge(local image store):"+
					" request-data=%#v",
				lsReq))
		return
	default:
		lsResListStr := slices.MapMust(lsResList, saya.LsResult.PlatformNameVersionTypeTaglike)
		resp.Diagnostics.AddError(
			"ImageResource -- Too many images found",
			fmt.Sprintf("ImageResource -- Too many images found: found=%v", lsResListStr))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	log.Debugf(ctx, "ImageResource.Read -- Update requested: request=%#v data=%#v", req, data)
	if resp.Diagnostics.HasError() {
		return
	}

	pullImgAndUpdateData(ctx, r, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.KeepLocally.ValueBool() {
		return
	}

	delReq := saya.ImageDeleteRequest{
		Name:     data.Name.ValueString(),
		ImgType:  data.ImgType.ValueString(),
		Platform: data.Platform.ValueString(),

		RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx(),
	}

	if err := saya.ImageRm(ctx, delReq); err != nil {
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}
}

func (r *ImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	idStr := req.ID
	id, err := saya.ParseImgId(idStr)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}

	lsReq := saya.LsRequest{
		Name:      id.R.Normalized(),
		ImgType:   id.ImgType,
		Platform:  id.P.PlatformStr(),
		OsVariant: "",
		Sha256:    "",

		RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx(),
	}

	switch lsResList, err := saya.Ls(ctx, lsReq); {
	case err != nil:
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	case len(lsResList) == 1:
		lsRes := lsResList[0]
		platformStr := lsRes.Platform.PlatformStr()
		id, err := saya.ImgId(lsRes.Name, lsRes.Version, platformStr, lsRes.Type)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
			return
		}
		data := &ImageResourceModel{}
		data.Id = types.StringValue(id)
		data.Sha256 = types.StringValue(lsRes.Sha256)
		data.Name = types.StringValue(lsRes.Name + ":" + lsRes.Version)
		data.ImgType = types.StringValue(lsRes.Type)
		data.Platform = types.StringValue(platformStr)
		data.RepoType = types.StringValue(lsRes.SrcType)
		data.HttpRepo = types.ObjectNull(attributesHttpRepo())
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	case len(lsResList) == 0:
		resp.Diagnostics.AddError(
			"ImageResource -- No image found in the forge",
			fmt.Sprintf(
				"ImageResource -- No image found in the forge(local image store):"+
					" request-data=%#v",
				lsReq))
		return
	default:
		lsResListStr := slices.MapMust(lsResList, saya.LsResult.PlatformNameVersionTypeTaglike)
		resp.Diagnostics.AddError(
			"ImageResource -- Too many images found",
			fmt.Sprintf("ImageResource -- Too many images found: found=%v", lsResListStr))
		return
	}

}
