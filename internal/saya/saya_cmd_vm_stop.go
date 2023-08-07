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

type CmdVmStop struct {
	SayaCmd
}

// CmdVmStop creates a saya cmd to stop a VM given its ID.
// The base command is <saya-exe> vm stop <id>
func NewCmdVmStop(id string, reqSayaCtx RequestSayaCtx) (*CmdVmStop, error) {
	sayaExe := strings.TrimSpace(reqSayaCtx.Exe)
	if sayaExe == "" {
		return nil, errors.Errorf("CmdVmStop -- saya-exe must not be blank")
	}

	cmd := &CmdVmStop{SayaCmd{exe: sayaExe}}
	cmd.Args.AppendNoValue("vm")
	cmd.Args.AppendNoValue("stop")
	if id = strings.TrimSpace(id); id == "" {
		return nil, errors.Errorf("NewCmdVmStop -- id must not be null")
	}

	cmd.Args.AppendNoValue(id)

	cmd.WithRequestSayaCtx(reqSayaCtx)

	return cmd, nil
}
