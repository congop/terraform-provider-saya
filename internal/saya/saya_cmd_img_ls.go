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

type SayaCmdLs struct {
	SayaCmd
}

func (cmd *SayaCmdLs) WithRef(tagName string) {
	tagName = strings.TrimSpace(tagName)
	cmd.Args.AppendNoValue(strings.TrimSpace(tagName))
}

// NewSayaCmdLs create a new image ls base command to be completed with flags.
// The base command is <saya-exe> image ls
func NewCmdImgLs(sayaExe string, reqSayaCtx RequestSayaCtx) (*SayaCmdLs, error) {
	if sayaExe = strings.TrimSpace(sayaExe); sayaExe == "" {
		return nil, errors.Errorf("NewSayaCmdLs -- saya-exe must not be blank")
	}
	cmd := &SayaCmdLs{SayaCmd{exe: sayaExe}}
	cmd.Args.AppendNoValue("image")
	cmd.Args.AppendNoValue("ls")

	cmd.WithRequestSayaCtx(reqSayaCtx)

	return cmd, nil
}
