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
	"encoding/json"
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

const EnvNameTfSayaForgeWithImg = "TF_SAYA_FORGE_WITH_IMG"
const EnvNameTfSayaTestAccVmResArgs = "TF_SAYA_TEST_ACC_VM_RES_ARGS"

func getForgeWithImgWebserverV1(t *testing.T) string {
	forge, avail := os.LookupEnv(EnvNameTfSayaForgeWithImg)
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

type tfTestAccVmResArgs struct {
	saya.Platform
	ImgType     string `json:"img_type,omitempty"`
	ComputeType string `json:"compute_type,omitempty"`
}

func (args *tfTestAccVmResArgs) ImgIdWebserverV1() string {
	// "linux/amd64:webserver:v1:qcow2"
	return strings.Join([]string{args.PlatformStr(), "webserver", "v1", args.ImgType}, ":")
}

func loadTfTestAccVmResArgs(t *testing.T) *tfTestAccVmResArgs {
	argsJsonStr, avail := os.LookupEnv(EnvNameTfSayaTestAccVmResArgs)
	require.Truef(t, avail,
		"args setting per environment variable expected: env-name=%s",
		EnvNameTfSayaTestAccVmResArgs)
	args := tfTestAccVmResArgs{}
	err := json.Unmarshal([]byte(argsJsonStr), &args)
	require.NoErrorf(t, err,
		"fail to json-parse tf test-acc-arg json from env: "+
			"\n\tenv-name=%s \n\tjson=%s",
		EnvNameTfSayaTestAccVmResArgs, argsJsonStr)
	return &args
}

func TestAccVmRes(t *testing.T) {
	os.Setenv("TF_ACC", "yes")
	os.Setenv("TF_ACC_LOG", "debug")
	os.Setenv("TF_LOG", "debug")
	if _, avail := os.LookupEnv(EnvNameTfSayaTestAccVmResArgs); !avail {
		os.Setenv(
			EnvNameTfSayaTestAccVmResArgs,
			`{"os":"linux", "arch":"arm64", "img_type":"qcow2", "compute_type":"qemu"}`)
	}

	args := loadTfTestAccVmResArgs(t)
	forge := getForgeWithImgWebserverV1(t)

	t.Cleanup(
		func() {
			rmVmByName(t, []string{"test1vm_read", "test1vm"}, forge)
		},
	)

	givenTest1VmInForgeFn := newGivenTest1VmInForgeFn(t, &repos.ImgInRepo{}, forge, args)

	imgIdWebserverV1 := args.ImgIdWebserverV1()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },

		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVmResourceConfig(t, &repos.ImgInRepo{}, forge, args),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_vm.test", "name", "test1vm"),
					resource.TestCheckResourceAttr("saya_vm.test", "image", imgIdWebserverV1),
					resource.TestCheckResourceAttr("saya_vm.test", "compute_type", args.ComputeType),
					resource.TestCheckResourceAttr("saya_vm.test", "os_variant", "alpine"),
					resource.TestCheckResourceAttr("saya_vm.test", "state", "running"),
					CheckTfVmResHasIdOfVmWIthName(t, "saya_vm.test", "test1vm", forge),
				),
				PreventPostDestroyRefresh: true,
			},

			{
				Config:            testAccVmResourceConfig(t, &repos.ImgInRepo{}, forge, args),
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
				Config: testAccVmResReadConfig(t, &repos.ImgInRepo{}, forge, args),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("saya_vm.test_read", "name", "test1vm_read"),
					resource.TestCheckResourceAttr("saya_vm.test_read", "image", imgIdWebserverV1),
					resource.TestCheckResourceAttr("saya_vm.test_read", "compute_type", args.ComputeType),
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

func testAccVmResourceConfig(
	t *testing.T, imgInRepo *repos.ImgInRepo, forge string,
	args *tfTestAccVmResArgs,
) string {
	tplStr := `
resource "saya_vm" "test" {
	name = "test1vm" 
	image = "{{ .args.ImgIdWebserverV1 }}"
	compute_type = "{{ .args.ComputeType }}"
	state = "running"
	# os_variant = "alpine"	
}`
	data := map[string]any{
		"imgInRepo": imgInRepo,
		"args":      args,
	}
	tpl := template.New("saya_vm_tf")
	tpl, err := tpl.Parse(tplStr)
	require.NoErrorf(t, err, "testAccVmResourceConfig-- fail to parse template")
	buf := bytes.Buffer{}
	err = tpl.Execute(&buf, data)
	require.NoErrorf(t, err, "testAccVmResourceConfig -- fail to execute template")

	return testProvider(t, imgInRepo, forge) + buf.String()

}

func newGivenTest1VmInForgeFn(t *testing.T, imgInRepo *repos.ImgInRepo, forge string, args *tfTestAccVmResArgs) func() string {
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
			ComputeType: args.ComputeType,   // "qemu",
			Platform:    args.PlatformStr(), // "linux/amd64",
			ImgType:     args.ImgType,       // "qcow2",
		})

		require.NoErrorf(t, err, "givenTest1VmInForge -- fail to create vm test1vm")
		id = runRes.Id
		return runRes.Id
	}
}

func testAccVmResReadConfig(
	t *testing.T, imgInRepo *repos.ImgInRepo, forge string,
	args *tfTestAccVmResArgs,
) string {

	tplStr := `
resource "saya_vm" "test_read" {
	name = "test1vm_read" 
	image = "{{ .args.ImgIdWebserverV1 }}"
	compute_type = "{{ .args.ComputeType }}"
	state = "stopped"
}`
	data := map[string]any{
		"imgInRepo": imgInRepo,
		"args":      args,
	}
	tpl := template.New("saya_vm_tf_read")
	tpl, err := tpl.Parse(tplStr)
	require.NoErrorf(t, err, "testAccVmResReadConfig-- fail to parse template")
	buf := bytes.Buffer{}
	err = tpl.Execute(&buf, data)
	require.NoErrorf(t, err, "testAccVmResReadConfig -- fail to execute template")

	return testProvider(t, imgInRepo, forge) + buf.String()

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
