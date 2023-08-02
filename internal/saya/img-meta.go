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
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/pkg/errors"
)

type ImageTagMetaData struct {
	Name      string            `yaml:"name" json:"name"`
	Version   string            `yaml:"version" json:"version"`
	Sha256    string            `yaml:"sha256" json:"sha256"`
	Type      string            `yaml:"type" json:"type"` // the image type e.g. ova, vhd, vmdk, iso. img, ...
	Platform  PlatformSw        `yaml:"platform" json:"platform"`
	CreatedAt time.Time         `yaml:"created_at" json:"created_at"`
	Labels    map[string]string `yaml:"labels" json:"labels"`
	SrcType   string            `yaml:"src_type,omitempty" json:"src_type,omitempty"` // build | convert | tag | http | s3
}

func ImageTagMetaDataFromJsonFile(ctx context.Context, jsonFilepath string) (*ImageTagMetaData, error) {
	fileFile, err := os.Open(jsonFilepath)
	if err != nil {
		return nil, errors.Wrapf(err,
			"ImageTagMetaDataFromJsonFile -- fail to open json file: \njson-file=%s \nerr=%s",
			jsonFilepath, err)
	}
	defer func() {
		if err := fileFile.Close(); err != nil {
			log.Warnf(ctx, "ImageTagMetaDataFromJsonFile -- fail to close json file: path=%s err=%v", jsonFilepath, err)
		}
	}()

	decoder := json.NewDecoder(fileFile)
	meta := ImageTagMetaData{}
	if err := decoder.Decode(&meta); err != nil {
		return nil, errors.Wrapf(err,
			"ImageTagMetaData -- fail to decode json: path=%s err=%v",
			jsonFilepath, err)
	}

	return &meta, err
}

// LsCmdRes model image in the forge returned as result of saya image ls command.
// e.g. [{"Created":"2023-07-03T08:45:36.599939836+02:00",
// "Id":"c87c9fac305744a229103177230db4192bdb6d4547c416ed32fc6bac9971befb",
// "ImgType":"qcow2","Name":"testeearm","OsVar":"alpine",
// "Platform":"linux/arm64","Size":1074135040,"Tag":"latest"}]
type LsCmdRes struct {
	Name      string `json:"name"`
	Version   string `json:"tag"`
	Sha256    string `json:"id"`
	Type      string `json:"img_type"` // the image type e.g. ova, vhd, vmdk, iso. img, ...
	Platform  string `json:"platform"`
	OsVariant string
	CreatedAt time.Time         `json:"created_at"`
	Labels    map[string]string `json:"labels"`

	SrcType string `json:"src_type,omitempty"` // build | convert | tag | http | s3
}

func ImageTagMetaDataListFromJsonFile(ctx context.Context, jsonFilepath string) ([]ImageTagMetaData, error) {
	fileFile, err := os.Open(jsonFilepath)
	if err != nil {
		return nil, errors.Wrapf(err,
			"ImageTagMetaDataFromJsonFile -- fail to open json file: \njson-file=%s \nerr=%s",
			jsonFilepath, err)
	}
	defer func() {
		if err := fileFile.Close(); err != nil {
			log.Warnf(ctx, "ImageTagMetaDataFromJsonFile -- fail to close json file: path=%s err=%v", jsonFilepath, err)
		}
	}()

	decoder := json.NewDecoder(fileFile)
	meta := []ImageTagMetaData{}
	if err := decoder.Decode(&meta); err != nil {
		// data, _ := os.ReadFile(jsonFilepath)
		return nil, errors.Wrapf(err,
			"ImageTagMetaData -- fail to decode json: path=%s err=%v",
			jsonFilepath, err)
	}

	return meta, err
}
