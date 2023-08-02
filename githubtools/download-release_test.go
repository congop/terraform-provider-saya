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
	"testing"

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
	type args struct {
		releaseName    string
		binInstallDir  string
		alreadyInstall bool
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
			name:    "should-override-old-installation",
			args:    args{releaseName: "", binInstallDir: t.TempDir(), alreadyInstall: true},
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
				"https://github.com/congop/saya-io/releases/download/saya-teaser-20230801T204622/saya-teaser-20230801T204622.zip",
				func(*http.Request) (*http.Response, error) {
					return httpmock.NewBytesResponse(200, sayaReleaseZipStub(t)), nil
				})
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
				theSayaExeInstalled(t, tt.args.binInstallDir)
			}
		})
	}
}
func givenSayaExeAlreadyInstalled(t *testing.T, binInstallDir string) {
	sayaExePath := filepath.Join(binInstallDir, "saya")
	err := os.WriteFile(sayaExePath, []byte("i-am-saya-exe-already-there-but-old"), 0600)
	require.NoErrorf(t, err, "fail to ensure that saya is install: err=%v", err)
}
func theSayaExeInstalled(t *testing.T, binInstallDir string) {
	expectedInstallSayaExePath := filepath.Join(binInstallDir, "saya")
	data, err := os.ReadFile(expectedInstallSayaExePath)
	require.NoErrorf(t, err,
		"could not read saya exe file at expected location: expectedLocation=%s err=%s",
		expectedInstallSayaExePath, err)
	require.Equalf(t,
		fakeSayaExeFileContent, string(data),
		"exe file content should have match expected")
	info, _ := os.Stat(expectedInstallSayaExePath)

	maskedExeUserGroupOther := info.Mode().Perm() & 0111
	require.Equalf(t, uint32(0111), uint32(maskedExeUserGroupOther),
		"saya exe should be executable for user, usergroup and other: expected=%s expected=%o actual=%o",
		info.Mode().Perm().String(), 0111, maskedExeUserGroupOther)
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
			Name: "saya-teaser-20230801T204622/bin/saya",
			Body: fakeSayaExeFileContent,
		},
		{
			Name: "saya-teaser-20230801T204622/saya-teaser-20230801T204622/THIRD-PARTY-NOTICES.txt",
			Body: "Saya uses the following opensource golang modules: \n ...",
		},
		{
			Name: "LICENSE",
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
