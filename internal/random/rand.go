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

package random

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/pkg/errors"
)

func String(lenBytes uint) (string, error) {
	buf, err := Bytes(lenBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func Bytes(lenBytes uint) ([]byte, error) {
	if lenBytes <= 0 {
		return nil, errors.Errorf("Bytes -- len-bytes must be greater 0")
	}

	buf := make([]byte, lenBytes)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, errors.Errorf("Bytes -- fail to run rand to fill bytes buf")
	}

	return buf, nil
}
