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
	"os"

	"github.com/congop/terraform-provider-saya/internal/log"
)

type VmRunRequest struct {
	RequestSayaCtx

	Name        string
	ImgRef      string
	ComputeType string
	Platform    string
	ImgType     string
}

type VmRunResult struct {
	Id        string
	Name      string
	OsVariant string
	Ssh       struct {
		Ip                  string
		Port                uint16
		User                string
		CreatedPwdFile      string
		CreatedIdentityFile string
	}
}

type VmRunResultCmdSsh struct {
	Ip                  string `json:"ip,omitempty"`
	Port                uint16 `json:"port,omitempty"`
	User                string `json:"user,omitempty"`
	CreatedPwdFile      string `json:"created_pwd_file,omitempty"`
	CreatedIdentityFile string `json:"created_identity_file,omitempty"`
}

// {"id":"81e4b6cf-ad66-409d-8a78-429b7499093e","name":"saya-2023-08-04T20-39-47Z-999d5a49cbbf","ssh":{"ip":"192.168.56.253","port":22,"user":"root"}}
type VmRunResultCmd struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	OsVariant string `json:"os_variant,omitempty"`

	Ssh *VmRunResultCmdSsh `json:"ssh,omitempty"`
}

func (cmdRes *VmRunResultCmd) ToVmRunResultCmd() VmRunResult {
	if cmdRes == nil {
		return VmRunResult{}
	}

	res := VmRunResult{
		Id:        cmdRes.Id,
		Name:      cmdRes.Name,
		OsVariant: cmdRes.OsVariant,
	}

	if cmdRes.Ssh != nil {
		res.Ssh.CreatedIdentityFile = cmdRes.Ssh.CreatedIdentityFile
		res.Ssh.CreatedPwdFile = cmdRes.Ssh.CreatedPwdFile
		res.Ssh.Ip = cmdRes.Ssh.Ip
		res.Ssh.Port = cmdRes.Ssh.Port
		res.Ssh.User = cmdRes.Ssh.User
	}
	return res
}

func VmRun(ctx context.Context, req VmRunRequest) (*VmRunResult, error) {
	cmd, err := NewCmdVmRun(req.ImgRef, req.RequestSayaCtx)
	if err != nil {
		return nil, err
	}
	cmd.appendFlagIfNotBlank("--name", req.Name)
	cmd.appendFlagIfNotBlank("--compute-type", req.ComputeType)
	cmd.appendFlagIfNotBlank("--platform", req.Platform)
	cmd.appendFlagIfNotBlank("--img-type", req.ImgType)

	resultDst, err := newTmpResultDstPath()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(resultDst); err != nil && !os.IsNotExist(err) {
			log.Warnf(ctx, "VmRun -- fail to remove ls result file: path=%s err=%v", resultDst, err)
		}
	}()

	cmd.appendFlagIfNotBlank("--result-dst", resultDst)

	outcome, err := cmd.Exec(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf(ctx, "VmRun -- saya vm run execution outcome: outcome=%#v", outcome)

	resCmdJson, err := DecodeJsonFile[VmRunResultCmd]("VmRun", ctx, resultDst)
	if err != nil {
		return nil, err
	}
	res := resCmdJson.ToVmRunResultCmd()
	return &res, nil
}
