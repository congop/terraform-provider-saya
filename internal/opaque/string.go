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

package opaque

import (
	"fmt"
	"strings"
)

// String wraps a string to avoid displaying it inadvertently.
//
// # Use Value() to get the actual string
//
// Note: It is not a simple alias to type string to prevent
// concatenation with string and conversion to string
type String struct {
	string
}

func NewString(value string) *String {
	return &String{string: value}
}

var _ fmt.Stringer = (*String)(nil)
var _ fmt.GoStringer = (*String)(nil)
var _ fmt.Formatter = (*String)(nil)

func (ostr String) String() string {
	if len(ostr.string) == 0 {
		return "<empty>"
	}
	return "********"
}

func (ostr String) Value() string {
	return ostr.string
}

func (ostr *String) SetValue(newVal string) {
	ostr.string = newVal
}

func (ostr *String) Normalize() {
	ostr.string = strings.TrimSpace(ostr.string)
}

func (ostr String) GoString() string {
	if len(ostr.string) == 0 {
		return "opaque.String(\"<empty>\")"
	}
	return "opaque.String(\"********\")"
}

func (ostr String) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		// we do not provide detail, e.g. for s.Flag('#')
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s", ostr.String())
	default:
		fmt.Fprintf(s, "%%!%c(%T=%s)", verb, ostr, ostr.String())
	}
}
