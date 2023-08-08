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
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/congop/terraform-provider-saya/internal/slices"
	"github.com/congop/terraform-provider-saya/internal/stringutil"
	repos "github.com/congop/terraform-provider-saya/internal/stubrepo"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

const TF_SAYA_FORGE_WITH_IMG_ENV = "TF_SAYA_FORGE_WITH_IMG"

func getForgeWithImgWebserverV1(t *testing.T) string {
	forge, avail := os.LookupEnv(TF_SAYA_FORGE_WITH_IMG_ENV)
	if avail {
		if forge = strings.TrimSpace(forge); forge != "" {
			require.DirExistsf(t, forge,
				"forge with image specified by environment variable must exists: path=%s",
				forge)
			return forge
		}

	}
	absPath, err := filepath.Abs("../../../vmbuilder/saya/.forge")
	require.NoErrorf(t, err, "fail to get absolute path of local forge with image")
	require.DirExistsf(t, absPath, "local forge with image must exists: abs-path=%s", absPath)
	return absPath
}

func rmVmByName(t *testing.T, vmNames []string, forge string) {
	errs := make([]error, 0, 8)
	for _, vmName := range vmNames {
		vms := findVmIdByName(t, vmName, forge)
		for _, vm := range vms {
			_, err := saya.VmRm(repos.StubLogCtx(), saya.VmRmRequest{
				RequestSayaCtx: saya.RequestSayaCtx{
					Forge: forge, Exe: sayaExe(t),
				},
				Id:   vm.Id,
				Name: vm.Name,
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) != 0 {
		// (error).Error instead of error.Error because vs-code auto-imports some core-os error module
		t.Fatalf(
			"rmVmByName: errs=%s",
			stringutil.IndentN(1, strings.Join(slices.MapMust(errs, (error).Error), "\n")),
		)
	}
}

func TestAccVmRes(t *testing.T) {
	os.Setenv("TF_ACC", "yes")
	os.Setenv("TF_ACC_LOG", "debug")
	os.Setenv("TF_LOG", "debug")
	forge := getForgeWithImgWebserverV1(t)

	t.Cleanup(
		func() {
			rmVmByName(t, []string{"test1vm_read", "test1vm"}, forge)
		},
	)

	givenTest1VmInForgeFn := newGivenTest1VmInForgeFn(t, &repos.ImgInRepo{}, forge)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },

		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVmResourceConfig(t, &repos.ImgInRepo{}, forge),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_vm.test", "name", "test1vm"),
					resource.TestCheckResourceAttr("saya_vm.test", "image", "linux/amd64:webserver:v1:qcow2"),
					resource.TestCheckResourceAttr("saya_vm.test", "compute_type", "qemu"),
					resource.TestCheckResourceAttr("saya_vm.test", "os_variant", "alpine"),
					resource.TestCheckResourceAttr("saya_vm.test", "state", "running"),
					CheckTfVmResHasIdOfVmWIthName(t, "saya_vm.test", "test1vm", forge),
				),
				PreventPostDestroyRefresh: true,
			},

			{
				Config:            testAccVmResourceConfig(t, &repos.ImgInRepo{}, forge),
				ResourceName:      "saya_vm.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return givenTest1VmInForgeFn(), nil
				},
				PreConfig: func() {
					givenTest1VmInForgeFn()
				},
			},

			// Update and Read testing
			{
				Config: testAccVmResReadConfig(t, &repos.ImgInRepo{}, forge),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_vm.test_read", "name", "test1vm_read"),
					resource.TestCheckResourceAttr("saya_vm.test_read", "image", "linux/amd64:webserver:v1:qcow2"),
					resource.TestCheckResourceAttr("saya_vm.test_read", "compute_type", "qemu"),
					resource.TestCheckResourceAttr("saya_vm.test_read", "os_variant", "alpine"),
					resource.TestCheckResourceAttr("saya_vm.test_read", "state", "stopped"),
					CheckTfVmResHasIdOfVmWIthName(t, "saya_vm.test_read", "test1vm_read", forge),
				),
			},

			// // Delete testing automatically occurs in TestCase
		},
	})
}

func CheckTfVmResHasIdOfVmWIthName(t *testing.T, tfResourceName string, name string, forge string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		id := findVmIdByNameMust(t, name, forge)
		resource.TestCheckResourceAttr(tfResourceName, "id", id)
		return nil
	}
}

func testAccVmResourceConfig(t *testing.T, imgInRepo *repos.ImgInRepo, forge string) string {
	tplStr := `
resource "saya_vm" "test" {
	name = "test1vm" 
	image = "linux/amd64:webserver:v1:qcow2"
	compute_type = "qemu"
	state = "running"
	# os_variant = "alpine"	
}`
	tpl := template.New("saya_vm_tf")
	tpl, err := tpl.Parse(tplStr)
	require.NoErrorf(t, err, "testAccVmResourceConfig-- fail to parse template")
	buf := bytes.Buffer{}
	err = tpl.Execute(&buf, imgInRepo)
	require.NoErrorf(t, err, "testAccVmResourceConfig -- fail to execute template")

	return testProvider(t, imgInRepo, forge) + buf.String()

}

func newGivenTest1VmInForgeFn(t *testing.T, imgInRepo *repos.ImgInRepo, forge string) func() string {
	id := ""

	return func() string {
		if id != "" {
			return id
		}

		const reqVmName = "test1vm"
		switch foundVms := findVmIdByName(t, reqVmName, forge); len(foundVms) {
		case 1:
			id = foundVms[0].Id
			return id
		case 0:
			break
		default:
			t.Fatalf(
				"found too many vm with given name while expecting at most 1: name=%s, vms=%#v",
				reqVmName, foundVms)
		}

		runRes, err := saya.VmRun(repos.StubLogCtx(), saya.VmRunRequest{
			RequestSayaCtx: saya.RequestSayaCtx{
				Forge: forge, Exe: sayaExe(t),
			},
			Name:        "test1vm",
			ImgRef:      "webserver:v1",
			ComputeType: "qemu",
			Platform:    "linux/amd64",
			ImgType:     "qcow2",
		})

		require.NoErrorf(t, err, "givenTest1VmInForge -- fail to create vm test1vm")
		id = runRes.Id
		return runRes.Id
	}
}

func testAccVmResReadConfig(t *testing.T, imgInRepo *repos.ImgInRepo, forge string) string {
	tplStr := `
resource "saya_vm" "test_read" {
	name = "test1vm_read" 
	image = "linux/amd64:webserver:v1:qcow2"
	compute_type = "qemu"
	state = "stopped"
}`
	return testProvider(t, imgInRepo, forge) + tplStr

}

func findVmIdByNameMust(t *testing.T, name string, forge string) string {
	sayaExe := sayaExe(t)

	vms, err := saya.VmLs(repos.StubLogCtx(), saya.VmLsRequest{
		Name: name,
		RequestSayaCtx: saya.RequestSayaCtx{
			Exe:   sayaExe,
			Forge: forge,
		},
	})
	require.NoErrorf(t, err, "fail to find vm with name: name=%s", name)
	require.Equalf(t, 1, len(vms), "exactly one vm expected: name=%s count=%d, vms=%#v", name, len(vms), vms)
	return vms[0].Id
}

func findVmIdByName(t *testing.T, name string, forge string) []saya.VmLsResult {
	sayaExe := sayaExe(t)

	vms, err := saya.VmLs(repos.StubLogCtx(), saya.VmLsRequest{
		Name: name,
		RequestSayaCtx: saya.RequestSayaCtx{
			Exe:   sayaExe,
			Forge: forge,
		},
	})
	require.NoErrorf(t, err, "fail to find vm with name: name=%s", name)
	return vms
}
