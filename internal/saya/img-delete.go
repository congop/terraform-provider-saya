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
	"github.com/congop/terraform-provider-saya/internal/opaque"
)

type ImageDeleteRequest struct {
	Name     string
	ImgType  string
	Platform string

	Exe        string        // saya executable command or path
	Config     string        // saya executable command or path
	Forge      string        // forge(local image store) location
	LicenseKey opaque.String // License key
	LogLevel   string        // log level error|warn|info|debug|trace
}

func ImageRm(ctx context.Context, req ImageDeleteRequest) error {
	log.Debugf(ctx, "ImageRm -- requested: request=%#v", req)
	ref, err := ParseReference(req.Name)
	if err != nil {
		return err
	}
	cmd, err := NewCmdImgDel(req.Exe)
	if err != nil {
		return err
	}
	cmd.WithRef(ref.Normalized())

	cmd.WithCfgFile(req.Config)
	cmd.WithForgeLocation(req.Forge)
	cmd.WithLicenseKey(req.LicenseKey)
	cmd.WithLogLevel(req.LogLevel)

	cmd.appendFlagIfNotBlank("--img-type", req.ImgType)
	cmd.appendFlagIfNotBlank("--platform", req.Platform)

	outcome, err := cmd.Exec(ctx)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "ImageRm -- outcome of saya image rm: outcome=%#v", outcome)

	return nil
}
