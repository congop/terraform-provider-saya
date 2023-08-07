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

package saya

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/slices"
	"github.com/pkg/errors"
)

type VmLsRequest struct {
	RequestSayaCtx

	Id          string
	Name        string
	ImgRef      string
	ComputeType string
	OsVariant   string
	State       string
}

type VmLsResult struct {
	Id          string
	Name        string
	Arch        string
	Os          string
	OsVariant   string
	BaseImg     string // format: webserver:v1:ova
	ComputeType string
	State       string
}

func (res *VmLsResult) ImgId() string {
	if res == nil {
		return "<!res-is-nil!>"
	}
	return res.Platform() + ":" + res.BaseImg
}

func (res *VmLsResult) Platform() string {
	if res == nil {
		return "<!res-is-nil-platform!>"
	}
	return res.Os + "/" + res.Arch
}

func (res *VmLsResult) LabelNameAndId() string {
	if res == nil {
		return "<!res-is-nil!>"
	}
	b := strings.Builder{}
	b.Grow(len(res.Name) + len(res.Id) + 8)
	b.WriteString(res.Name)
	b.WriteRune('(')
	b.WriteString(res.Id)
	b.WriteRune(')')
	return b.String()
}

type VmLsResultCmd struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Arch        string `json:"arch"`
	Os          string `json:"os"`
	OsVariant   string `json:"os_variant"`
	BaseImg     string `json:"base_img"` // format: webserver:v1:ova
	ComputeType string `json:"compute_type"`
	Status      string `json:"status"`
}

func (resJson *VmLsResultCmd) ToVmLsResult() VmLsResult {
	if resJson == nil {
		return VmLsResult{}
	}

	return VmLsResult{
		Id:          resJson.Id,
		Name:        resJson.Name,
		Arch:        resJson.Arch,
		Os:          resJson.Os,
		OsVariant:   resJson.OsVariant,
		BaseImg:     resJson.BaseImg,
		ComputeType: resJson.ComputeType,
		State:       resJson.Status, // TODO check status mapping
	}
}

func VmLs(ctx context.Context, req VmLsRequest) ([]VmLsResult, error) {
	sayaCmd, err := NewCmdVmLs(req.Id, req.RequestSayaCtx)
	if err != nil {
		return nil, err
	}

	resultDst, err := newTmpResultDstPath()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(resultDst); err != nil && !os.IsNotExist(err) {
			log.Warnf(ctx, "VmLs -- fail to remove ls result file: path=%s err=%v", resultDst, err)
		}
	}()
	sayaCmd.appendFlagIfNotBlank("--format", "json")
	sayaCmd.appendFlagIfNotBlank("--result-dst", resultDst)
	// TODO add other flags available filter [base-image id label compute-type os-variant arch status]
	// , but having id result in at most 1 vm, so we are fine for the moment
	// even if this function s not a full ls command

	singleValueFilters := map[string]string{
		"name":        req.Name,
		"compte-type": req.ComputeType,
		"os-variant":  req.OsVariant,
	}
	filtersStr := make([]string, 0, len(singleValueFilters))
	for k, v := range singleValueFilters {

		if v = strings.TrimSpace(v); v != "" {
			filtersStr = append(filtersStr, k+"="+v)
		}
	}
	sayaCmd.appendMultiFlagIfNotEmpty("--filter", filtersStr)

	outcome, err := sayaCmd.Exec(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf(ctx, "VmLs -- saya vm ls execution outcome: outcome=%#v", outcome)

	resListJson, err := DecodeJsonFile[[]VmLsResultCmd]("VmLs", ctx, resultDst) // vmLsResultFromJson(ctx, resultDst)
	if err != nil {
		return nil, err
	}

	resList := slices.MapPMust(*resListJson, (*VmLsResultCmd).ToVmLsResult)

	return resList, nil
}

func DecodeJsonFile[T any](usecaseCtx string, ctx context.Context, jsonFilepath string) (*T, error) {
	fileFile, err := os.Open(jsonFilepath)
	var model *T
	if err != nil {
		return model, errors.Wrapf(err,
			"DecodeJsonFile -- fail to open json file: \n\tctx=%s \njson-file=%s \nerr=%s",
			usecaseCtx, jsonFilepath, err)
	}
	defer func() {
		if err := fileFile.Close(); err != nil {
			log.Warnf(ctx,
				"DecodeJsonFile -- fail to close json file: "+
					"\n\tctx=%s  \n\tpath=%s \n\terr=%+v",
				usecaseCtx, jsonFilepath, err)
		}
	}()

	decoder := json.NewDecoder(fileFile)
	if err := decoder.Decode(&model); err != nil {
		return model, errors.Wrapf(err,
			"DecodeJsonFile -- fail to decode json: \n\tctx=%s  \n\tpath=%s \n\terr=%v",
			usecaseCtx, jsonFilepath, err)
	}

	return model, err
}
