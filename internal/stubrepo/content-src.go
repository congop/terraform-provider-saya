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
	"bytes"
	"io"
)

func ContentSrcString(label, content string) func() (_body io.ReadCloser, _length int64, _err error) {
	return func() (_body io.ReadCloser, _length int64, _err error) {
		buf := bytes.NewBufferString(content)
		return io.NopCloser(buf), int64(buf.Len()), nil
	}
}

func ContentSrcBytes(label string, content []byte) func() (_body io.ReadCloser, _length int64, _err error) {
	return func() (_body io.ReadCloser, _length int64, _err error) {
		buf := bytes.NewBuffer(content)
		return io.NopCloser(buf), int64(buf.Len()), nil
	}
}
