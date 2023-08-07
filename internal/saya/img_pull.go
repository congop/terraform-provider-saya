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
	"path/filepath"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/opaque"
	"github.com/congop/terraform-provider-saya/internal/random"
)

type PullRequest struct {
	Name     string
	ImgType  string
	Platform string
	Hash     string
	RepoType string
	HttpRepo *HttpRepo

	Exe        string        // saya executable command or path
	Config     string        // saya executable command or path
	Forge      string        // forge(local image store) location
	LicenseKey opaque.String // License key
	LogLevel   string        // log level error|warn|info|debug|trace
}

type PullResult struct {
	Name     string
	Version  string
	Sha256   string
	Type     string // the image type e.g. ova, vhd, vmdk, iso. img, ...
	Platform PlatformSw
	SrcType  string // type of the repo used to fetch the image; e.g. http, s3
}

func Pull(ctx context.Context, req PullRequest) (*PullResult, error) {
	log.Debugf(ctx, "Pull requested: request=%#v", req)
	ref, err := ParseReference(req.Name)
	if err != nil {
		return nil, err
	}
	cmd, err := NewCmdImgPull(req.Exe)
	if err != nil {
		return nil, err
	}
	cmd.WithRef(ref.Normalized())

	cmd.WithCfgFile(req.Config)
	cmd.WithForgeLocation(req.Forge)
	cmd.WithLicenseKey(req.LicenseKey)
	cmd.WithLogLevel(req.LogLevel)

	cmd.appendFlagIfNotBlank("--hash", req.Hash)
	cmd.appendFlagIfNotBlank("--img-type", req.ImgType)
	cmd.appendFlagIfNotBlank("--platform", req.Platform)
	cmd.appendFlagIfNotBlank("--repo-type", req.RepoType)

	//cmd.appendFlagIfNotBlank("--repo-type", req.RepoType)

	if repo := req.HttpRepo; repo != nil {
		cmd.appendFlagIfNotBlank("--http-auth-basic-password", repo.AuthHttpBasic.Pwd)
		cmd.appendFlagIfNotBlank("--http-auth-basic-username", repo.AuthHttpBasic.Username)
		cmd.appendFlagIfNotBlank("--http-base-path", repo.BasePath)
		cmd.appendFlagIfNotBlank("--http-repo-url", repo.RepoUrl)
		cmd.appendFlagIfNotBlank("--http-upload-strategy", repo.UploadStrategy)
	}

	resultDst, err := newTmpResultDstPath()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(resultDst); err != nil && !os.IsNotExist(err) {
			log.Warnf(ctx, "Pull -- fail to remove pull result file: path=%s err=%v", resultDst, err)
		}
	}()
	cmd.appendFlagIfNotBlank("--result-dst", resultDst)

	outcome, err := cmd.Exec(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf(ctx, "Pull -- saya image pull execution outcome: outcome=%#v\n", outcome)
	meta, err := ImageTagMetaDataFromJsonFile(ctx, resultDst)
	if err != nil {
		return nil, err
	}

	// saya.ExecOutcome{
	//	Stdout:"7:16PM INF Pull -- image pulled: linux/arm64:ubuntu:v1:ova#472c9bc5d01d2f98b102d5d6d3477d6145b4326c694e6990213239395f9f3e4d\n"
	res := PullResult{
		Sha256:   meta.Sha256,
		Name:     meta.Name,
		Version:  meta.Version,
		Type:     meta.Type,
		Platform: meta.Platform,
		SrcType:  meta.SrcType,
	}
	log.Debugf(ctx, "Pull -- pull result: \n\toutcome=%#v \n\tresult=%#v", outcome, res)
	return &res, nil
}

// newTmpResultDstPath return a random new file path in the tmp directory, note the no file is created.
func newTmpResultDstPath() (string, error) {
	basePart, err := random.String(8)
	if err != nil {
		return "", err
	}
	base := fmt.Sprintf("saya-tf-%s.json", basePart)
	resultDst := filepath.Join(os.TempDir(), base)
	return resultDst, nil
}
