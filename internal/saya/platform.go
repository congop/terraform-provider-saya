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

	"github.com/containerd/containerd/platforms"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type Platform struct {
	Os          string `json:"os,omitempty"`
	Arch        string `json:"arch,omitempty"`
	ArchVariant string `json:"arch_variant,omitempty"`
}

func (platform Platform) ArchWithVariant() string {
	archVar := strings.TrimSpace(platform.ArchVariant)
	if archVar == "" {
		return platform.Arch
	}
	return platform.Arch + "/" + archVar
}

func (platform Platform) PlatformSegmentParts() []string {
	if archVar := platform.ArchVariant; archVar != "" {
		return []string{platform.Os, platform.Arch, archVar}
	}
	return []string{platform.Os, platform.Arch}
}

// PlatformSegment returns a / separated str representation os[/arch[/arch-variant]].
//
// e.g linux/amd64, linux/arm64, linux/arm/v7
func (platform Platform) PlatformStr() string {
	return strings.Join(platform.PlatformSegmentParts(), "/")
}

// PlatformSw is a platform which knows the os variant.
// It is more software oriented because the os-variant hints at what kind of software natively belongs.
// e.g. apt for ubuntu but rpm for redhat
type PlatformSw struct {
	Platform

	OsVariant string `json:"os_variant,omitempty"`
}

// PlatformNormalized return a normalized form of the specified platform string.
// It defaults to the current host platform, if passed platform string is blank.
func PlatformNormalized(
	platformStr string,
) (*Platform, error) {
	platformStr = strings.TrimSpace(platformStr)

	var platform v1.Platform
	var err error
	if platformStr == "" {
		platform = platforms.DefaultSpec()
	} else {
		platform, err = platforms.Parse(platformStr)
		if err != nil {
			err := errors.Wrapf(err,
				"PlatformNormalized -- fail to parse patform string, platform-str=%s",
				platformStr)
			return nil, err
		}
	}
	pNormalized := platforms.Normalize(platform)
	p := Platform{
		Os:          pNormalized.OS,
		Arch:        pNormalized.Architecture,
		ArchVariant: pNormalized.Variant,
	}
	return &p, nil
}
