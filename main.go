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
	"context"
	"flag"
	"log"

	"github.com/congop/terraform-provider-saya/internal/loglevel"
	"github.com/congop/terraform-provider-saya/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Formats terraform file in examples directory
// @requires terraform executable in path
//go:generate terraform fmt -recursive ./examples/

// Generates plugin documentation
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// Set at build time
	// @by  goreleaser
	// @see https://goreleaser.com/cookbooks/using-main.version/
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/congop/saya",
		Debug:   debug,
	}

	sayaLogLevel := loglevel.FromEnv()

	err := providerserver.Serve(context.Background(), provider.New(version, sayaLogLevel), opts)

	if err != nil {
		log.Fatalf("%+v", err)
	}
}
