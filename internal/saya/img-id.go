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
	"strings"

	"github.com/congop/terraform-provider-saya/internal/slices"
	"github.com/pkg/errors"
)

// ImgId  return an id which can be used to identify a tagged image in a repository.
// The format is <platform>:<name>:<version>:<imgType>
// where:
//
// <platform> = os/arch[os-variant] ;
//
// name = the name of the image
//
// version = the version or tag of the image
//
// imgType = the type of the image (e.g. ova, vmdk, img, qcow2, ...)
func ImgId(name, version, platform, imgType string) (string, error) {
	parts := []string{platform, name, version, imgType}
	partsNormalized := slices.FilterMust(
		slices.MapMust(parts, strings.TrimSpace),
		StringNotEmpty)

	if len(partsNormalized) != len(parts) {
		return "", errors.Errorf("ImgId -- illegal attribute, all must be non blank:"+
			"name=%q version=%q platform=%q imgType=%q",
			name, version, platform, imgType)
	}
	return strings.Join(partsNormalized, ":"), nil
}

func StringNotEmpty(str string) bool {
	return len(str) > 0
}

type ParsedImgId struct {
	P       Platform
	R       Reference
	ImgType string
}

func ParseImgId(str string) (*ParsedImgId, error) {
	parts := strings.Split(strings.TrimSpace(str), ":")
	parts = slices.FilterMust(slices.MapMust(parts, strings.TrimSpace), StringNotEmpty)
	if len(parts) != 4 {
		return nil, errors.Errorf(
			"ParseImgId -- illegal format: expected=platform:name:version:type, got=%q parts=%q",
			str, parts)
	}
	platformStr := parts[0]
	p, err := PlatformNormalized(platformStr)
	if len(parts) != 4 {
		return nil, errors.Wrapf(err,
			"ParseImgId -- fail to parse platform string: expected=os/arch[/arv-variant], "+
				"platform-str=%s img-id=%q parts=%q \n\terr=%s",
			platformStr, str, parts, err)
	}
	piId := ParsedImgId{
		P:       *p,
		R:       ReferenceFromNameAndVersion(parts[1], parts[2]),
		ImgType: parts[3],
	}
	return &piId, nil
}
