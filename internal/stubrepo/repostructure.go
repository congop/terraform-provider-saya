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

package stubrepo

import (
	"strings"

	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/pkg/errors"
)

var supportedTypes = []string{"ova", "vmdk", "vhd", "img", "iso", "qcow2", "vdi"}

func ImgFileName(imgType string) (string, error) {
	imgType = strings.TrimSpace(imgType)
	for _, supported := range supportedTypes {
		if supported == imgType {
			return "img." + imgType, nil
		}
	}
	return "", errors.Errorf("image type(%s) is not supported; supported are %s", imgType, supportedTypes)
}

func ImageRepoDirRelUrlSegments(platform *saya.Platform, name string, version string) []string {
	versionNormalized := strings.TrimSpace(version)
	if versionNormalized == "" {
		versionNormalized = "latest"
	}
	platformSegs := platform.PlatformSegmentParts()
	segs := make([]string, 0, 2+len(platformSegs))
	segs = append(segs, name, versionNormalized)
	segs = append(segs, platformSegs...)
	return segs
}

func ImageRepoRelUrlSegments(
	baseDirSegs []string, platform *saya.Platform,
	name string, version string, imgType string,
) ([]string, error) {
	urlPathSegments := ImageRepoDirRelUrlSegments(platform, name, version)
	imageFileBasename, err := ImgFileName(imgType)
	if err != nil {
		return nil, err
	}
	segs := make([]string, 0, len(baseDirSegs)+len(urlPathSegments)+1)
	for _, baseDirSeg := range baseDirSegs {
		baseDirSeg = strings.TrimSpace(baseDirSeg)
		if baseDirSeg != "" {
			segs = append(segs, baseDirSeg)
		}
	}
	segs = append(segs, urlPathSegments...)
	segs = append(segs, imageFileBasename)
	return segs, nil
}
