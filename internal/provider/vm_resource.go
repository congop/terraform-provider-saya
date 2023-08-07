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
	"time"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/poll"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/congop/terraform-provider-saya/internal/slices"
	"github.com/congop/terraform-provider-saya/internal/stringutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VmResource{}
var _ resource.ResourceWithImportState = &VmResource{}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

// VmResource defines the resource implementation.
type VmResource struct {
	sayaExeCtx *SayaExecutionCtx
}

// VmResourceModel describes the resource data model.
type VmResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Image       types.String `tfsdk:"image"`
	ComputeType types.String `tfsdk:"compute_type"`
	Id          types.String `tfsdk:"id"`
	State       types.String `tfsdk:"state"` // steady state of the vm; started | stopped (what about hibernate e.g. for aws)
	OsVariant   types.String `tfsdk:"os_variant"`

	KeepOnDelete types.Bool `tfsdk:"keep_on_delete"` // true if vm is not to be delete even vm double in terraform state will
}

func (r *VmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	log.Debugf(ctx, "VmResource.MetaData=%#v\n\n", req)
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

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
			"os_variant": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "os-variant, e.g. alpine, ubuntu, debian",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "vm name",
				Description:         "vm name",
				Optional:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "image id, format: platform:name:version:image-type",
				Required:            true,
			},
			"compute_type": schema.StringAttribute{
				MarkdownDescription: "compute-type, e.g. qemu, virtualbox",
				Description:         "compute-type, e.g. qemu, virtualbox",
				Optional:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "state state of the vm, choices: started, stopped",
				Description:         "state state of the vm, choices: started, stopped",
				Optional:            true,
			},
			"keep_on_delete": schema.BoolAttribute{
				MarkdownDescription: "true if vm is not to be delete even vm double in terraform state will",
				Description:         "true if vm is not to be delete even vm double in terraform state will",
				Optional:            true,
			},
		},
	}

}

func (r *VmResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *VmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	imgId, err := saya.ParseImgId(data.Image.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}

	pullReq := saya.VmRunRequest{
		Name:           data.Name.ValueString(),
		ImgRef:         imgId.R.Normalized(),
		Platform:       imgId.P.PlatformStr(),
		ImgType:        imgId.ImgType,
		ComputeType:    data.ComputeType.ValueString(),
		RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx(),
	}

	pullRes, err := saya.VmRun(ctx, pullReq)
	if err != nil {
		if pullRes != nil && pullRes.Id != "" {
			if err := deleteVm("VmResource.Create-CleanupOnVmRunFailed", ctx, pullRes.Id, r.sayaExeCtx.ToRequestSayaCtx()); err != nil {
				log.Warnf(ctx, "%+v", err)
			}
		}
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}

	id := pullRes.Id
	data.Id = types.StringValue(id)

	if state := strings.TrimSpace(data.State.ValueString()); state == "stopped" {
		poller := poll.Poller[*saya.VmStopResult]{
			Interval:             time.Second * 5,
			Timeout:              time.Minute * 5,
			MaxConsecutiveErrors: 2, // we effectively will try max twice, since not having an error means success
			OutcomeNillable:      true,
			ConditionFunc: func(outcome *saya.VmStopResult) (bool, error) {
				return outcome != nil, nil
			},
			LastOutcome: nil,
			OutcomeGetter: func() (*saya.VmStopResult, error) {
				return saya.VmStop(ctx, saya.VmStopRequest{Id: id, RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx()})
			},
		}
		if err := poller.Poll(ctx); err != nil {
			// TODO better message in case same ctx-issue (e.g. canceling) caused poll exit
			if err := deleteVm("VmResource.Create-EnsureStateStoppedFailed", ctx, pullRes.Id, r.sayaExeCtx.ToRequestSayaCtx()); err != nil {
				log.Warnf(ctx, "%+v", err)
			}
			resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
			return
		}
	}

	log.Tracef(ctx, "Create -- saya vm tf-created: runRes=%#v", pullRes)
	data.OsVariant = types.StringValue(pullRes.OsVariant)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func deleteVm(actionCtx string, ctx context.Context, id string, sayaExeCtx saya.RequestSayaCtx) error {
	// cleaning up

	if _, err := saya.VmStop(ctx, saya.VmStopRequest{Id: id, RequestSayaCtx: sayaExeCtx}); err != nil {
		log.Debugf(ctx,
			"VmResource.Create -- fail to stop vm: action-ctx=%s id=%s err=%s",
			actionCtx, id, stringutil.IndentN(2, fmt.Sprintf("%+v", err)))
	}
	if _, err := saya.VmRm(ctx, saya.VmRmRequest{Id: id, RequestSayaCtx: sayaExeCtx}); err != nil {
		return err
	}
	return nil
}

func (r *VmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *VmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateVmResWithVmDataById(ctx, r, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	log.Debugf(ctx, "VmResource.Read -- Update requested: request=%#v data=%#v", req, data)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO this may be too easy to be true; test what update really means, tf-model change --?--> remove old vm start new one?
	updateVmResWithVmDataById(ctx, r, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *VmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.KeepOnDelete.ValueBool() {
		return
	}

	id := data.Id.ValueString()

	if err := deleteVm("VmResource.Delete", ctx, id, r.sayaExeCtx.ToRequestSayaCtx()); err != nil {
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	}
}

func (r *VmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	idStr := req.ID

	lsReq := saya.VmLsRequest{
		Id:             idStr,
		RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx(),
	}

	switch lsResList, err := saya.VmLs(ctx, lsReq); {
	case err != nil:
		resp.Diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	case len(lsResList) == 1:
		lsRes := lsResList[0]
		data := &VmResourceModel{}
		data.Id = types.StringValue(idStr)
		data.Name = types.StringValue(lsRes.Name)
		data.Image = types.StringValue(lsRes.ImgId())
		data.ComputeType = types.StringValue(lsRes.ComputeType)
		data.OsVariant = types.StringValue(lsRes.OsVariant)
		data.State = types.StringValue(lsRes.State)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	case len(lsResList) == 0:
		resp.Diagnostics.AddError(
			"VmResource -- No vm found in the forge",
			fmt.Sprintf(
				"VmResource -- No vm found in the forge(local image store):"+
					" request-data=%#v",
				lsReq))
		return
	default:
		lsResListStr := slices.MapPMust(lsResList, (*saya.VmLsResult).LabelNameAndId)
		resp.Diagnostics.AddError(
			"VmResource -- Too many vm found",
			fmt.Sprintf("VmResource -- Too many vm found: found=%v", lsResListStr))
		return
	}

}

func updateVmResWithVmDataById(ctx context.Context, r *VmResource, data *VmResourceModel, diagnostics *diag.Diagnostics) {
	idStr := data.Id.ValueString()
	lsReq := saya.VmLsRequest{
		Id:             idStr,
		RequestSayaCtx: r.sayaExeCtx.ToRequestSayaCtx(),
	}

	switch lsResList, err := saya.VmLs(ctx, lsReq); {
	case err != nil:
		diagnostics.AddError(err.Error(), fmt.Sprintf("%+v", err))
		return
	case len(lsResList) == 1:
		lsRes := lsResList[0]
		data.Id = types.StringValue(idStr)
		data.Name = types.StringValue(lsRes.Name)
		data.Image = types.StringValue(lsRes.ImgId())
		data.ComputeType = types.StringValue(lsRes.ComputeType)
		data.OsVariant = types.StringValue(lsRes.OsVariant)
		data.State = types.StringValue(lsRes.State)
		return
	case len(lsResList) == 0:
		diagnostics.AddError(
			"updateVmResWithVmDataById -- No vm found in the forge",
			fmt.Sprintf(
				"updateVmResWithVmDataById -- No vm found in the forge(local image store):"+
					" request-data=%#v",
				lsReq))
		return
	default:
		lsResListStr := slices.MapPMust(lsResList, (*saya.VmLsResult).LabelNameAndId)
		diagnostics.AddError(
			"updateVmResWithVmDataById -- Too many vm found",
			fmt.Sprintf("updateVmResWithVmDataById -- Too many vm found: found=%v", lsResListStr))
		return
	}
}
