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

import "github.com/congop/terraform-provider-saya/internal/opaque"

type RequestSayaCtx struct {
	Exe        string        // saya executable command or path
	Config     string        // saya executable command or path
	Forge      string        // forge(local image store) location
	LicenseKey opaque.String // License key
	LogLevel   string        // log level error|warn|info|debug|trace

}
