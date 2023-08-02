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

package loglevel

import (
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

func tfLogLevel() string {
	for _, key := range []string{"TF_LOG_PROVIDER", "TF_LOG"} {
		val, avail := os.LookupEnv(key)
		if !avail {
			continue
		}
		if val = strings.TrimSpace(val); val != "" {
			return val
		}
	}
	return ""
}

// FromEnv return saya tf log level setting from environment.
func FromEnv() string {
	logLevel := strings.ToLower(tfLogLevel())
	if logLevel == "" {
		return ""
	}

	if !slices.Contains([]string{"error", "warn", "info", "debug", "trace"}, logLevel) {
		logLevel = "trace"
	}

	return logLevel
}
