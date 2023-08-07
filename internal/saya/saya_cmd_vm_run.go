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
	"strings"

	"github.com/pkg/errors"
)

type CmdVmRun struct {
	SayaCmd
}

// CmdVmRun creates a saya cmd to run a VM.
// The base command is <saya-exe> vm run <img-reference>
func NewCmdVmRun(imgRef string, reqSayaCtx RequestSayaCtx) (*CmdVmRun, error) {
	sayaExe := strings.TrimSpace(reqSayaCtx.Exe)
	if sayaExe == "" {
		return nil, errors.Errorf("CmdVmRun -- saya-exe must not be blank")
	}

	cmd := &CmdVmRun{SayaCmd{exe: sayaExe}}
	cmd.Args.AppendNoValue("vm")
	cmd.Args.AppendNoValue("run")
	imgRef = strings.TrimSpace(imgRef)
	if imgRef == "" {
		return nil, errors.Errorf("CmdVmRun -- image-reference must not be blank")
	}
	cmd.Args.AppendNoValue(imgRef)

	cmd.WithRequestSayaCtx(reqSayaCtx)

	return cmd, nil
}
