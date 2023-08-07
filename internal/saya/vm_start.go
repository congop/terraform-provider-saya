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

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/stringutil"
	"github.com/pkg/errors"
)

type VmStartRequest struct {
	Id string

	RequestSayaCtx
}

type VmStartResult struct {
	Id string
}

func VmStart(ctx context.Context, req VmStartRequest) (*VmStartResult, error) {
	cmd, err := NewCmdVmStart(req.Id, req.RequestSayaCtx)
	if err != nil {
		return nil, err
	}
	outcome, err := cmd.Exec(ctx)
	if err != nil {
		return nil, errors.Wrapf(err,
			"VmStart --  fail to execute saya vm start command: "+
				"\n\treq=%#v \n\terr=%s  ",
			req, stringutil.IndentN(2, err.Error()))

	}
	log.Debugf(ctx, "VmStart -- cmd exec outcome: outcome=%#v", outcome)
	return &VmStartResult{Id: req.Id}, nil
}
