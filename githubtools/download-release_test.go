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

package githubtools

import (
	"archive/zip"
	"bytes"
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/congop/execstub"
	"github.com/congop/execstub/pkg/comproto"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

//go:embed sample-github-releases-api-resp.json
var statics embed.FS

func getTestReleaseResp(t *testing.T) []byte {
	content, err := statics.ReadFile("sample-github-releases-api-resp.json")
	require.NoErrorf(t, err,
		"fail to get static resource: sample-github-releases-api-resp.json: err=%v",
		err)
	return content
}

func TestInstallSayaReleaseBin(t *testing.T) {
	stubber := execstub.NewExecStubber()
	installCmd := ""
	runNopExit0 := func(sreq comproto.StubRequest) *comproto.ExecOutcome {
		installCmd = strings.Join(sreq.Args, " ")
		return &comproto.ExecOutcome{Key: sreq.Key, Stdout: "", ExitCode: 0}
	}

	defer stubber.CleanUp()
	settings := comproto.Settings{
		Mode:     comproto.StubbingModeDyna,
		ExecType: comproto.ExecTypeExe,
	}
	_, err := stubber.WhenExecDoStubFunc(runNopExit0, "sudo", settings)
	require.NoErrorf(t, err, "fail to stub sudo command")
	altReleaseUrl := "http://localhost:1235/dasas/dasasa"

	type args struct {
		releaseName   string
		binInstallDir string
		altReleaseUrl *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "should-install-saya-bin",
			args:    args{releaseName: "", binInstallDir: t.TempDir()},
			wantErr: false,
		},
		{
			name:    "should-install-saya-bin-by-using-alternative-url-env-spec",
			args:    args{releaseName: "", binInstallDir: t.TempDir(), altReleaseUrl: &altReleaseUrl},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", sayaIoReleaseJsonUrl,
				httpmock.NewBytesResponder(200, getTestReleaseResp(t)))
			httpmock.RegisterResponder(
				"GET",
				"https://github.com/congop/saya-io/releases/download/saya_teaser-20231005T135240/saya_teaser-20231005T135240_linux_amd64.zip",
				func(*http.Request) (*http.Response, error) {
					return httpmock.NewBytesResponse(200, sayaReleaseZipStub(t)), nil
				})
			require.NoErrorf(t, os.Unsetenv("SAYA_RELEASE_URL"), "fail to unset environment variable SAYA_RELEASE_URL")
			if tt.args.altReleaseUrl != nil {
				httpmock.RegisterResponder(
					"GET",
					*tt.args.altReleaseUrl,
					func(*http.Request) (*http.Response, error) {
						return httpmock.NewBytesResponse(200, sayaReleaseZipStub(t)), nil
					})
				os.Setenv("SAYA_RELEASE_URL", *tt.args.altReleaseUrl)
			}
			// https://github.com
			httpmock.RegisterResponder(
				"GET", `=~^https://.*github\.com.*\z`,
				httpmock.NewStringResponder(404, `no available yet`))
			givenSayaExeAlreadyInstalled(t, tt.args.binInstallDir)

			// when
			err := InstallSayaReleaseBin(tt.args.releaseName, tt.args.binInstallDir)

			//then
			if (err != nil) != tt.wantErr {
				t.Errorf("InstallSayaReleaseBin() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				expectedCmdParts := []string{
					"/saya", "setup",
					"--compute-type", "localhost", "--target", "localhost",
					"--want-compute-type", "qemu",
					"--target-user", "runner",
					"--log-level", "info",
				}
				expectedCmdSuffix := strings.Join(expectedCmdParts, " ")
				require.Contains(t, installCmd, expectedCmdSuffix,
					"should have been actual install cmd suffix: \n\tinstall-cmd=%s \n\texpected-cmd-suffix=%s",
					installCmd, expectedCmdSuffix)
				//theSayaExeInstalled(t, tt.args.binInstallDir)
			}
		})
	}
}
func givenSayaExeAlreadyInstalled(t *testing.T, binInstallDir string) {
	sayaExePath := filepath.Join(binInstallDir, "saya")
	err := os.WriteFile(sayaExePath, []byte("i-am-saya-exe-already-there-but-old"), 0600)
	require.NoErrorf(t, err, "fail to ensure that saya is install: err=%v", err)
}

const fakeSayaExeFileContent = "#!/bin/bash\n echo 'fake saya cmd'\n exit 1"

func sayaReleaseZipStub(t *testing.T) []byte {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
	}{
		{
			Name: "saya_teaser-20231005T135240_linux_amd64/bin/saya",
			Body: fakeSayaExeFileContent,
		},
		{
			Name: "saya_teaser-20231005T135240_linux_amd64/THIRD-PARTY-NOTICES/THIRD-PARTY-NOTICES.txt",
			Body: "Saya uses the following opensource golang modules: \n ...",
		},
		{
			Name: "saya_teaser-20231005T135240_linux_amd64/LICENSE",
			Body: "commercial.",
		},
	}
	for _, file := range files {
		f, err := w.Create(file.Name)

		require.NoErrorf(t, err,
			"sayaReleaseZipStub -- fail to create zip entry file: name=%s",
			file.Name)
		_, err = f.Write([]byte(file.Body))
		require.NoErrorf(t, err,
			"sayaReleaseZipStub -- fail to write zip entry file content: name=%s",
			file.Name)
	}

	// Make sure to check the error on Close.
	err := w.Close()
	require.NoErrorf(t, err,
		"sayaReleaseZipStub -- fail to close  zip stream: err:=%v",
		err)
	return buf.Bytes()
}
