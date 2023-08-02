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

	"github.com/pkg/errors"
)

type Reference struct {
	Original string // the original reference value, e.g. ubuntu:latest.
	Name     string // name
	Version  string // Normalized i.e.  latest if None was given
}

func (ref Reference) Normalized() string {
	return ref.Name + ":" + ref.Version
}

func ReferenceFromNameAndVersion(name, version string) Reference {
	return Reference{
		Original: name + ":" + version,
		Name:     name,
		Version:  version,
	}
}

func ParseReference(ref string) (*Reference, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, errors.Errorf("ParseReference -- reference must not be blank")
	}
	parts := strings.Split(ref, ":")
	switch len(parts) {
	case 1:
		return &Reference{Original: ref, Name: strings.TrimSpace(parts[0]), Version: "latest"}, nil
	case 2:
		return &Reference{Original: ref, Name: strings.TrimSpace(parts[0]), Version: strings.TrimSpace(parts[1])}, nil
	default:
		return nil, errors.Errorf("ParseReference -- bad format: expected=name[:tag], got=%s", ref)
	}
}
