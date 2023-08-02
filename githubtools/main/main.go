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

package main

import (
	"flag"
	"log"

	"github.com/congop/terraform-provider-saya/githubtools"
)

func main() {
	var binInstallDir string
	var releaseName string

	flag.StringVar(&binInstallDir, "bin-install-dir", "/usr/bin/", "set the target saya binary installation directory; the default is /usr/bin")
	flag.StringVar(&releaseName, "release-name", "", "set the name of the saya release to download")

	flag.Parse()

	err := githubtools.InstallSayaReleaseBin(releaseName, binInstallDir)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
