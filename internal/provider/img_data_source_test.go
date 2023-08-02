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
	"testing"

	"github.com/congop/terraform-provider-saya/internal/stubrepo"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccExampleDataSource(t *testing.T) {
	os.Setenv("TF_ACC", "yes")
	stubrepo.WithStubLogCtxLevel("debug")

	baseDir := t.TempDir()
	forge, err := os.MkdirTemp(baseDir, "forge")
	require.NoErrorf(t, err, "fail to create forge directory: base-dir=%s", baseDir)
	ubuntuVDataread1OvaLinuxArmInForge := givenUbuntuVxOvaLinuxArmInForge(t, forge, "v_data_read_1")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccImgDataRead(t, forge),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.saya_image.readtest", "name", "ubuntu:v_data_read_1"),
					resource.TestCheckResourceAttr("data.saya_image.readtest", "img_type", "ova"),
					resource.TestCheckResourceAttr("data.saya_image.readtest", "platform", "linux/arm64"),
					resource.TestCheckResourceAttr("data.saya_image.readtest",
						"sha256", ubuntuVDataread1OvaLinuxArmInForge.Sha256),
				),
			},
		},
	})
}

func testAccImgDataRead(t *testing.T, forge string) string {
	return testProvider(t, nil, forge) + `
data "saya_image" "readtest" {
  name = "ubuntu:v_data_read_1"
  img_type = "ova"
  platform = "linux/arm64"
  filters = {
	label = ["audience=tester"]
  }
}
`
}
