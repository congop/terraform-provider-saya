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

package provider

import (
	"os"
	"os/exec"
	"testing"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/congop/terraform-provider-saya/internal/opaque"
	"github.com/congop/terraform-provider-saya/internal/saya"
	repos "github.com/congop/terraform-provider-saya/internal/stubrepo"
	"github.com/stretchr/testify/require"
)

func sayaExe(t *testing.T) string {
	sayaExe, avail := os.LookupEnv("SAYA_EXE")
	if !avail {
		sayaInPath, err := exec.LookPath("saya")
		if err == nil {
			return sayaInPath
		}

		// this is my owner local setup
		return "../../../vmbuilder/saya/saya"
	}
	return sayaExe
}

// givenUbuntuVxOvaLinuxArmInForge ensures an image with the given specification is in the local db.
// Note that a http repository is started and closed. But the corresponding http repo data are set.
func givenUbuntuVxOvaLinuxArmInForge(
	t *testing.T, forge string, version string,
) *repos.ImgInRepo {

	sayaExe := sayaExe(t)
	specData := repos.PullImgSpecData{
		Tag:       saya.Reference{Original: "ubuntu:" + version, Name: "ubuntu", Version: version},
		Hash:      "",
		OsVariant: "ubuntu",
		Platform:  saya.Platform{Os: "linux", Arch: "arm64"},
		RepoType:  "http",
		ImgType:   "ova",
	}
	httpRepo := repos.NewDummyHttpRepo()
	require.NoErrorf(t, httpRepo.Start(), "fail to start http server")
	defer func() {
		if err := httpRepo.Close(); err != nil {
			log.Errorf(repos.StubLogCtx(), "fail to stop http server: err=%+v ", err)
		}
	}()
	img, err := repos.GivenImgInRemoteRepo(httpRepo.RegisterDummyImg, specData, true)
	require.NoErrorf(t, err, "fail to put image in forge")
	pullReq := saya.PullRequest{
		Name:     specData.Tag.Normalized(),
		ImgType:  specData.ImgType,
		Platform: specData.Platform.PlatformStr(),
		Hash:     "",
		RepoType: "http",
		HttpRepo: httpRepo.AsRepos().Http,
		Exe:      sayaExe,
		Forge:    forge,
	}
	_, err = saya.Pull(repos.StubLogCtx(), pullReq)
	require.NoErrorf(t, err, "fail to pull image into forge")
	img.Repos = httpRepo.AsRepos()
	return img
}

func givenUbuntuV1OvaLinuxArmInHttpRepo(dummyRepos *repos.RemoteRepos) (*repos.ImgInRepo, error) {
	specData := repos.PullImgSpecData{
		Tag:       saya.Reference{Original: "ubuntu:v1", Name: "ubuntu", Version: "v1"},
		Hash:      "",
		OsVariant: "ubuntu",
		Platform:  saya.Platform{Os: "linux", Arch: "arm64"},
		RepoType:  "http",
		ImgType:   "ova",
	}

	img, err := repos.GivenImgInRemoteRepo(dummyRepos.Http.RegisterDummyImg, specData, true)
	if err != nil {
		return nil, err
	}
	img.Repos = dummyRepos.Http.AsRepos()
	return img, nil
}

func givenImageIsInForgeByPullingFn(
	t *testing.T, imgIdStr string, remoteRepos *repos.RemoteRepos, forge string,
) func() {
	return func() {
		sayaExe := sayaExe(t)
		imgId, err := saya.ParseImgId(imgIdStr)
		require.NoErrorf(t, err, "failed to parse image-id: imgId=%s", imgIdStr)
		req := saya.PullRequest{
			Name:       imgId.R.Normalized(),
			ImgType:    imgId.ImgType,
			Platform:   imgId.P.PlatformStr(),
			Hash:       "",
			RepoType:   "http",
			HttpRepo:   remoteRepos.Http.AsRepos().Http,
			Exe:        sayaExe,
			Forge:      forge,
			Config:     "",
			LicenseKey: *opaque.NewString(""),
			LogLevel:   "debug",
		}
		t.Logf("in-http repo = %v", remoteRepos.Http.RepoContent())

		_, err = saya.Pull(repos.StubLogCtx(), req)
		require.NoErrorf(t, err, "failed to pull")

		res, err := saya.Ls(repos.StubLogCtx(), saya.LsRequest{
			RequestSayaCtx: saya.RequestSayaCtx{
				Forge: forge, Exe: sayaExe,
			},
		})
		t.Logf("ls-results = %#v  \n\terr=%v", res, err)
	}
}
