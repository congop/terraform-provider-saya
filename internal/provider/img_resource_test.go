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
	"bytes"
	"os"
	"testing"
	"text/template"

	repos "github.com/congop/terraform-provider-saya/internal/stubrepo"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccImageRes(t *testing.T) {
	os.Setenv("TF_ACC", "yes")
	// os.Setenv("TF_ACC_TERRAFORM_VERSION", "v1.5.3")
	os.Setenv("TF_ACC_LOG", "debug")
	os.Setenv("TF_LOG", "debug")
	baseDir := t.TempDir()
	forge, err := os.MkdirTemp(baseDir, "forge")
	require.NoErrorf(t, err, "fail to create forge directory: base-dir=%s", baseDir)

	dummyRepos := repos.RemoteRepos{}
	require.NoError(t, dummyRepos.GivenHttpRepoStarted(), "fail to create and start dummy http repo")
	require.NoError(t, dummyRepos.GivenS3RepoStarted(), "fail to create and start dummy s3 repo")
	ubuntuV1OvaLinuxArmInHttpRepo, err := givenUbuntuV1OvaLinuxArmInHttpRepo(&dummyRepos)
	require.NoErrorf(t, err, "fail to create linux/arm64:ubuntu:v1:ova in http repo")
	t.Logf("ubuntuV1OvaLinuxArmInHttpRepo = %#v", ubuntuV1OvaLinuxArmInHttpRepo)
	ubuntuVdataread1OvaLinuxArmInForge := givenUbuntuVxOvaLinuxArmInForge(t, forge, "read_v1")
	t.Cleanup(func() {
		if t.Failed() {
			repo := dummyRepos.S3
			repo.Log(t.Logf)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },

		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//  Create repoType=s3
			{
				Config: testAccExampleResourceConfig(t, ubuntuV1OvaLinuxArmInHttpRepo, forge, "s3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_image.test", "name", "ubuntu:v1"),
					resource.TestCheckResourceAttr("saya_image.test", "img_type", "ova"),
					resource.TestCheckResourceAttr("saya_image.test", "platform", "linux/arm64"),
					resource.TestCheckResourceAttr("saya_image.test", "id", "linux/arm64:ubuntu:v1:ova"),
					resource.TestCheckResourceAttr("saya_image.test", "sha256", ubuntuV1OvaLinuxArmInHttpRepo.Sha256),
				),

				PreventPostDestroyRefresh: true,
			},
			// Create and Read testing - repo-tape=http
			{
				Config: testAccExampleResourceConfig(t, ubuntuV1OvaLinuxArmInHttpRepo, forge, "http"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_image.test", "name", "ubuntu:v1"),
					resource.TestCheckResourceAttr("saya_image.test", "img_type", "ova"),
					resource.TestCheckResourceAttr("saya_image.test", "platform", "linux/arm64"),
					resource.TestCheckResourceAttr("saya_image.test", "id", "linux/arm64:ubuntu:v1:ova"),
					resource.TestCheckResourceAttr("saya_image.test", "sha256", ubuntuV1OvaLinuxArmInHttpRepo.Sha256),
				),
				PreventPostDestroyRefresh: true,
			},

			// ImportState testing
			// testStepNewImportState
			// https://github.com/hashicorp/terraform-plugin-sdk/commit/ba4b604a523c06119692cdc12d43af35e2a80008
			// https://github.com/hashicorp/terraform-plugin-sdk/blob/df85539d6ddceae8203a2b8b0d7c127e88e133af/helper/resource/testing_new_import_state.go
			{
				Config:            testAccExampleResourceConfig(t, ubuntuV1OvaLinuxArmInHttpRepo, forge, "http"),
				ResourceName:      "saya_image.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "linux/arm64:ubuntu:v1:ova",
				PreConfig: givenImageIsInForgeByPullingFn(
					t, "linux/arm64:ubuntu:v1:ova", &dummyRepos, forge,
				),
			},
			// Update and Read testing
			{
				Config: testAccImgResRead(t, ubuntuVdataread1OvaLinuxArmInForge, forge),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_image.readtest", "name", "ubuntu:read_v1"),
					resource.TestCheckResourceAttr("saya_image.readtest", "img_type", "ova"),
					resource.TestCheckResourceAttr("saya_image.readtest", "platform", "linux/arm64"),
					resource.TestCheckResourceAttr("saya_image.readtest", "sha256", ubuntuVdataread1OvaLinuxArmInForge.Sha256),
				),
			},
			// // Delete testing automatically occurs in TestCase
		},
	})
}

func testProvider(t *testing.T, imgInRepo *repos.ImgInRepo, forge string) string {
	sayaExe := sayaExe(t)
	data := struct {
		*repos.ImgInRepo
		Forge   string
		SayaExe string
	}{
		ImgInRepo: imgInRepo,
		Forge:     forge,
		SayaExe:   sayaExe,
	}
	tplStr := `
	provider "saya" {
	  version = "0.0.1"
	  exe = "{{ .SayaExe }}"
	  forge = "{{ .Forge }}"
	  {{- if .HasHttpRepo}}
        http_repo = {
          url = "{{ .Repos.Http.RepoUrl }}"
          base_path = "{{ .Repos.Http.BasePath }}"
        }
      {{- end }}
	  {{- if .HasS3Repo}}
        s3_repo = {
          bucket = "{{ .Repos.S3.Bucket }}"
          base_key = "{{ .Repos.S3.BaseKey }}"
		  region =  "{{ .Repos.S3.Region }}"
		  ep_url_s3 =  "{{ .Repos.S3.EpUrlS3 }}"
		  use_path_style = true
		  {{- if .HasS3RepoCredentials}}
		    {{ $Creds := .Repos.S3.AuthAwsCreds }}
			credentials = {
				access_key_id = "{{ $Creds.AccessKeyID }}"
				secret_access_key  = "{{ $Creds.SecretAccessKey }}"
				session_token  = "{{ $Creds.SessionToken }}"
				source  = "{{ $Creds.Source}}"
				CanExpire =  "{{ $Creds.CanExpire }}"
				{{- if  $Creds.HasExpires }}
					Expires = "{{ $Creds.Expires }}"
				{{- end }}
			}
		  {{- end }}
        }
      {{- end }}
    }
	`
	tpl := template.New("saya_provider_tf")
	tpl, err := tpl.Parse(tplStr)
	require.NoErrorf(t, err, "testProvider-- fail to parse template")
	buf := bytes.Buffer{}
	err = tpl.Execute(&buf, data)
	require.NoErrorf(t, err, "testProvider -- fail to execute template")
	return buf.String()
}

func testAccExampleResourceConfig(t *testing.T, imgInRepo *repos.ImgInRepo, forge string, repoType string) string {
	tplStr := `
resource "saya_image" "test" {
	name = "ubuntu:v1" 
	img_type = "ova"
	platform = "linux/arm64"
	repo_type = "{{ .repoType }}"	
}`
	tpl := template.New("saya_image_tf")
	tpl, err := tpl.Parse(tplStr)
	require.NoErrorf(t, err, "testAccExampleResourceConfig-- fail to parse template")
	buf := bytes.Buffer{}
	data := map[string]any{
		"imgInRepo": imgInRepo,
		"repoType":  repoType,
	}
	err = tpl.Execute(&buf, data)
	require.NoErrorf(t, err, "testAccExampleResourceConfig -- fail to execute template")

	return testProvider(t, imgInRepo, forge) + buf.String()

}

func testAccImgResRead(t *testing.T, imgInRepo *repos.ImgInRepo, forge string) string {
	return testProvider(t, imgInRepo, forge) + `
resource "saya_image" "readtest" {
  name = "ubuntu:read_v1"
  img_type = "ova"
  platform = "linux/arm64"
  repo_type = "http"
  #filters = {
  #	label = ["audience=tester"]
  #}
}
`
}
