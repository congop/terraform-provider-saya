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
	"fmt"
	"os"
	"strings"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/opaque"
)

type LsRequest struct {
	Name      string
	ImgType   string
	Platform  string
	OsVariant string
	Sha256    string
	// Hash     string
	// RepoType string
	// HttpRepo *HttpRepo

	Exe        string        // saya executable command or path
	Config     string        // saya executable command or path
	Forge      string        // forge(local image store) location
	LicenseKey opaque.String // License key
	LogLevel   string        // log level error|warn|info|debug|trace
}

type LsResult struct {
	Name     string
	Version  string
	Sha256   string
	Type     string // the image type e.g. ova, vhd, vmdk, iso. img, ...
	Platform PlatformSw
	SrcType  string
}

func (res LsResult) PlatformNameVersionTypeTaglike() string {
	pSeg := res.Platform.PlatformStr()
	return strings.Join([]string{pSeg, res.Name, res.Version, res.Type}, ":")
}

func Ls(ctx context.Context, req LsRequest) ([]LsResult, error) {
	log.Debugf(ctx, "Ls -- requested: request=%#v", req)
	refNormalized := ""
	if name := strings.TrimSpace(req.Name); name != "" {
		ref, err := ParseReference(req.Name)
		if err != nil {
			return nil, err
		}
		refNormalized = ref.Normalized()
	}
	cmd, err := NewSayaCmdLs(req.Exe)
	if err != nil {
		return nil, err
	}
	if refNormalized != "" {
		cmd.WithRef(refNormalized)
	}

	cmd.WithCfgFile(req.Config)
	cmd.WithForgeLocation(req.Forge)
	cmd.WithLicenseKey(req.LicenseKey)
	cmd.WithLogLevel(req.LogLevel)

	// cmd.appendFlagIfNotBlank("--hash", req.Hash)
	// saya image ls --filter img-type=qcow2 --filter reference='appserver:v1*'
	// --filter os-variant=alpine --filter 'label=audience=tester'
	filters := make([]string, 0, 8)
	if imgType := strings.TrimSpace(req.ImgType); imgType != "" {
		filters = append(filters, fmt.Sprintf("img-type=%s", imgType))
	}

	// @customizable this could depends of the runtime host, if we proceed with empty platform
	if platformStr := strings.TrimSpace(req.Platform); platformStr != "" {
		platform, err := PlatformNormalized(platformStr)
		if err != nil {
			return nil, err
		}

		filters = append(filters, fmt.Sprintf("os=%s", platform.Os), fmt.Sprintf("arch=%s", platform.ArchWithVariant()))
	}
	cmd.appendMultiFlagIfNotEmpty("--filter", filters)

	resultDst, err := newTmpResultDstPath()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(resultDst); err != nil && !os.IsNotExist(err) {
			log.Warnf(ctx, "Ls -- fail to remove ls result file: path=%s err=%v", resultDst, err)
		}
	}()
	cmd.appendFlagIfNotBlank("--format", "json")
	cmd.appendFlagIfNotBlank("--result-dst", resultDst)

	outcome, err := cmd.Exec(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf(ctx, "Ls -- saya image ls execution outcome: outcome=%#v\n", outcome)
	metaList, err := ImageTagMetaDataListFromJsonFile(ctx, resultDst)
	if err != nil {
		return nil, err
	}

	if len(metaList) == 0 {
		return nil, nil
	}

	results := make([]LsResult, 0, len(metaList))
	for _, meta := range metaList {
		res := LsResult{
			Sha256:   meta.Sha256,
			Name:     meta.Name,
			Version:  meta.Version,
			Type:     meta.Type,
			Platform: meta.Platform,
			SrcType:  meta.SrcType,
		}
		results = append(results, res)
	}
	log.Debugf(ctx, "Ls -- saya image ls result: meta-lists=%#v results=%#v\n", metaList, results)
	return results, nil
}
