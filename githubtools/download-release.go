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
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	sayaSlices "github.com/congop/terraform-provider-saya/internal/slices"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func InstallSayaReleaseBin(releaseName string, binInstallDir string) error {
	releaseUrl := ""

	altReleaseUrl, avail := os.LookupEnv("SAYA_RELEASE_URL")
	if avail {
		log.Printf("InstallSayaReleaseBin -- using saya release url setting from environment: altReleaseUrl=%s", altReleaseUrl)
		releaseUrl = altReleaseUrl
	} else {
		altReleaseUrl, err := getReleaseUrl(releaseName)
		if err != nil {
			return err
		}
		releaseUrl = altReleaseUrl
	}

	tmpDir, err := os.MkdirTemp("", "dld-saya-release-*")
	if err != nil {
		return errors.Wrapf(err, "InstallSayaReleaseBin -- fail to create tempdir to download saya release: err=%v", err)
	}

	releaseZipPath, err := func() (string, error) {
		// "browser_download_url": "https://github.com/congop/saya-io/releases/download/saya-teaser-20230801T204622/saya-teaser-20230801T204622.zip"
		log.Printf("InstallSayaReleaseBin -- getting saya release zip: at=%s", releaseUrl)
		resp, err := http.Get(releaseUrl)
		if err != nil {
			return "", errors.Wrapf(err, "InstallSayaReleaseBin -- fail to get release: url=%s err=%v", releaseUrl, err)
		}
		defer resp.Body.Close()

		nameBse := path.Base(releaseUrl)
		fileTarget, err := os.CreateTemp(tmpDir, nameBse)
		if err != nil {
			return "", errors.Wrapf(err, "InstallSayaReleaseBin -- fail to create file to save download: "+
				"dir=%s, base-name=%s url=%s err=%v", tmpDir, nameBse, releaseUrl, err)
		}
		defer fileTarget.Close()
		_, err = io.Copy(fileTarget, resp.Body)
		if err != nil {
			return "", errors.Wrapf(err, "InstallSayaReleaseBin -- fail to copy get body to file: "+
				"dir=%s, base-name=%s url=%s err=%v", tmpDir, nameBse, releaseUrl, err)
		}

		return fileTarget.Name(), nil
	}()
	if err != nil {
		return err
	}

	if err := extractSayaBinary(releaseZipPath, binInstallDir, tmpDir); err != nil {
		return err
	}
	return nil
}

func extractSayaBinary(releaseZipPath string, binInstallDir string, tmpDir string) error {

	r, err := zip.OpenReader(releaseZipPath)
	if err != nil {
		return errors.Wrapf(err, "extractSayaBinary -- fail to open saya distribution zip: dist-path=%s err=%s", releaseZipPath, err)
	}
	defer r.Close()

	var sayaExeFileExtractedPath string

	entries := make([]string, 0, len(r.File))

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/bin/saya") {
			sayaExeFileExtractedPath, err = func(f *zip.File) (string, error) {
				rc, err := f.Open()
				if err != nil {
					return "", errors.Wrapf(err,
						"extractSayaBinary -- fail to open saya distribution zip entry: dist-path=%s entry=%s err=%s",
						releaseZipPath, f.Name, err)
				}
				defer rc.Close()

				sayaExeFile, err := os.Create(filepath.Join(tmpDir, "saya"))
				if err != nil {
					return "", errors.Wrapf(err, "extractSayaBinary -- fail to create file to extract binary: err=%s", err)
				}

				defer sayaExeFile.Close()

				_, err = io.Copy(sayaExeFile, rc)
				if err != nil {
					return "", errors.Wrapf(err, "extractSayaBinary -- fail to copy saya executable data: err=%s", err)
				}
				return sayaExeFile.Name(), nil
			}(f)
			if err != nil {
				return err
			}
			break
		}
		entries = append(entries, f.Name)
	}

	if sayaExeFileExtractedPath == "" {
		return errors.Errorf("extractSayaBinary -- saya binary not found: entry-path=xxx/bin/saya , entries=%q", entries)
	}

	// "rwxrwxrwx"
	// sample of terraform: -rwxr-xr-x 1 root root 65110016 Jul 26 20:16 /usr/bin/terraform*

	// finalDestination := filepath.Join(binInstallDir, "saya")
	// err = func() error {
	// 	srcFile, err := os.Open(sayaExeFileExtractedPath)
	// 	if err != nil {
	// 		return errors.Wrapf(err,
	// 			"extractSayBinary -- failed to open extracted saya exe: path=%s",
	// 			sayaExeFileExtractedPath)
	// 	}
	// 	defer srcFile.Close()
	// 	dstFile, err := os.OpenFile(finalDestination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	// 	if err != nil {
	// 		return errors.Wrapf(err,
	// 			"extractSayBinary -- failed to open saya exe final destination: dst=%s",
	// 			finalDestination)
	// 	}
	// 	defer dstFile.Close()
	// 	if _, err := io.Copy(dstFile, srcFile); err != nil {
	// 		return errors.Wrapf(err,
	// 			"extractSayBinary -- failed to copy saya exe to destination: dst=%s src=%s",
	// 			finalDestination, sayaExeFileExtractedPath)
	// 	}
	// 	return nil
	// }()

	// if err != nil {
	// 	return err
	// }

	// if err := os.Chmod(finalDestination, 0755); err != nil {
	// 	return errors.Wrapf(err,
	// 		"extractSayBinary -- fail to file mode to 0755: finalDestination=%s",
	// 		finalDestination)
	// }
	if err := runSayaSetup(sayaExeFileExtractedPath); err != nil {
		return err
	}
	return nil
}

func runSayaSetup(sayaExeFileExtractedPath string) error {
	if err := os.Chmod(sayaExeFileExtractedPath, 0755); err != nil {
		return errors.Wrapf(err,
			"runSayaSetupVirtualBox -- fail to file mode to 0755: sayaExeFileExtractedPath=%s err=%v",
			sayaExeFileExtractedPath, err)
	}
	cmd := exec.Command("sudo", sayaExeFileExtractedPath, "setup",
		"--compute-type", "localhost", "--target", "localhost",
		"--want-compute-type", "qemu",
		"--target-user", "runner",
		"--log-level", "info",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err,
			"runSayaSetupVirtualBox -- fail to run setup: sayaExeFileExtractedPath=%s err=%v",
			sayaExeFileExtractedPath, err)
	}
	return nil
}

const releaseEntryKeyName = "name"

func getReleaseUrl(releaseName string) (string, error) {
	releasesJsonBytes, err := getSayaIoReleases()
	if err != nil {
		return "", err
	}
	releaseJson := []map[string]any{}
	if err := json.Unmarshal(releasesJsonBytes, &releaseJson); err != nil {
		return "", errors.Wrapf(err, "getReleaseUrl -- fail to parse releaseJsonUrl: err=%v, json=%s",
			err, string(releasesJsonBytes))
	}

	var foundRelease map[string]any

	switch releaseName = strings.TrimSpace(releaseName); {
	case len(releaseJson) == 0:
		return "", errors.Errorf("getReleaseUrl -- not found (json is empty)")
	case releaseName != "":
		idx := slices.IndexFunc(releaseJson, func(r map[string]any) bool {
			name, avail, err := MapValue[string, string](r, releaseEntryKeyName)
			switch {
			case err != nil:
				log.Printf("getReleaseUrl -- fail to get name: available-keys=%v err=%+v", maps.Keys(r), err)
				return false
			case !avail:
				return false
			default:
				return name == releaseName
			}
		})
		if idx == -1 {
			availNames := sayaSlices.MapMust(releaseJson, func(m map[string]any) any {
				name, avail := m[releaseEntryKeyName]
				if !avail {
					name = "<no-entry-with-key-name>"
				}
				return name
			})
			return "", errors.Errorf(
				"getReleaseUrl -- not found: wanted-name=%s availablename=%q",
				releaseEntryKeyName, availNames)
		}
		foundRelease = releaseJson[idx]
	default:
		slices.SortFunc[[]map[string]any](releaseJson, func(a, b map[string]any) int {
			aName := getPublishedAt(a)
			bName := getPublishedAt(b)
			switch {
			case aName < bName:
				return +1
			case aName > bName:
				return -1
			default:
				return 0

			}
		})
		foundRelease = releaseJson[0]
	}
	assets, avail, err := MapArrayValue[string, map[string]any](foundRelease, "assets")
	if err != nil {
		return "", errors.Wrapf(err,
			"getReleaseUrl -- error while getting assets: "+
				"\n\twanted-name=assets \n\tavailablename=%q \n\terr=%v",
			maps.Keys(foundRelease), err)
	}
	if !avail {
		return "", errors.Errorf(
			"getReleaseUrl -- assets not found: availablename=%q",
			maps.Keys(foundRelease))
	}
	asset := assets[0]
	browserUrl, avail, err := MapValue[string, string](asset, "browser_download_url")
	if err != nil {
		return "", errors.Wrapf(err,
			"getReleaseUrl -- error while getting browser url from asset: wanted-name=browser_download_url availablename=%q",
			maps.Keys(asset))
	}
	if !avail {
		return "", errors.Errorf(
			"getReleaseUrl -- browser url not found: availablename=%q",
			maps.Keys(asset))
	}
	return browserUrl, nil
}

func getReleaseName(m map[string]any) string {
	name, avail := m[releaseEntryKeyName]
	if !avail {
		panic(errors.Errorf("getReleaseName -- name entry not found: available-keys=%q", maps.Keys(m)))
	}
	nameStr, ok := name.(string)
	if !ok {
		panic(errors.Errorf("getReleaseName -- name entry with unexpected type: wanted-type=string actual-type=%T", name))
	}
	return nameStr
}

func getPublishedAt(m map[string]any) string {
	name, avail := m["published_at"]
	if !avail {
		panic(errors.Errorf("getPublishedAt -- published_at entry not found: available-keys=%q", maps.Keys(m)))
	}
	nameStr, ok := name.(string)
	if !ok {
		panic(errors.Errorf("getPublishedAt -- published_at entry with unexpected type: wanted-type=string actual-type=%T", name))
	}
	return nameStr
}

const sayaIoReleaseJsonUrl = "https://api.github.com/repos/congop/saya-io/releases"

func getSayaIoReleases() ([]byte, error) {
	log.Printf("getSayaIoReleases-- geting saya-io releases json: at %s", sayaIoReleaseJsonUrl)
	resp, err := http.Get(sayaIoReleaseJsonUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "getSayaIoReleases -- fail to get releases url: url=%s err=%v", sayaIoReleaseJsonUrl, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("getSayaIoReleases -- failed to close body: err=%v", err)
		}
	}()
	jsonBody, err := io.ReadAll(resp.Body)
	return jsonBody, err
}

func MapValue[K comparable, T any](m map[K]any, key K) (T, bool, error) {
	value, avail := m[key]
	if !avail {
		var valueRet T
		return valueRet, false, nil
	}
	valueConverted, ok := value.(T)
	if !ok {
		// should be "nil" value
		return valueConverted, true, errors.Errorf(
			"MapValue -- value cannot be converted to wanted type: "+
				"\n\twantedType=%T \n\tactualType=%T \n\tvalue=%#v",
			valueConverted, value, value)
	}

	return valueConverted, true, nil
}

func MapArrayValue[K comparable, T any](m map[K]any, key K) ([]T, bool, error) {
	value, avail := m[key]
	if !avail {
		var valueRet []T
		return valueRet, false, nil
	}

	valueIntermediate, ok := value.([]any)
	if !ok {
		var valueRet []T
		return valueRet, true, errors.Errorf(
			"MapValue -- value cannot be converted to intermediate type: "+
				"\n\twantedType=%T \n\tactualType=%T \n\tvalue=%#v",
			valueIntermediate, value, value)
	}
	valueRet := make([]T, 0, len(valueIntermediate))
	for _, vi := range valueIntermediate {
		vr, ok := vi.(T)
		if !ok {
			return valueRet, true, errors.Errorf(
				"MapValue -- value cannot be converted to ret type: "+
					"\n\twantedType=%T \n\tactualType=%T \n\tvalue=%#v",
				vr, vi, vi)
		}
		valueRet = append(valueRet, vr)
	}

	return valueRet, true, nil
}
