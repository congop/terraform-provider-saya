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

type CmdVmLs struct {
	SayaCmd
}

// NewCmdVmLs create a saya cmd to list VMs with the given ID.
// The base command is <saya-exe> vm ls <id-if-not-blank>
func NewCmdVmLs(id string, reqSayaCtx RequestSayaCtx) (*CmdVmLs, error) {
	sayaExe := strings.TrimSpace(reqSayaCtx.Exe)
	if sayaExe == "" {
		return nil, errors.Errorf("NewCmdVmLs -- saya-exe must not be blank")
	}

	cmd := &CmdVmLs{SayaCmd{exe: sayaExe}}
	cmd.Args.AppendNoValue("vm")
	cmd.Args.AppendNoValue("ls")
	if id = strings.TrimSpace(id); id != "" {
		cmd.Args.AppendNoValue(id)
	}

	cmd.WithRequestSayaCtx(reqSayaCtx)

	return cmd, nil
}
