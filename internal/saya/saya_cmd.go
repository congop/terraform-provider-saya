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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	stderrors "errors"

	"github.com/congop/terraform-provider-saya/internal/opaque"
	"github.com/congop/terraform-provider-saya/internal/stringutil"
	"github.com/pkg/errors"
)

type SayaCmd struct {
	exe                  string
	Args                 CmdArgs
	ArgsValidationErrors []error
}

func (sayaCmd *SayaCmd) WithRequestSayaCtx(req RequestSayaCtx) {
	sayaCmd.WithCfgFile(req.Config)
	sayaCmd.WithForgeLocation(req.Forge)
	sayaCmd.WithLicenseKey(req.LicenseKey)
	sayaCmd.WithLogLevel(req.LogLevel)
}

func (sayaCmd *SayaCmd) WithCfgFile(cfg string) {
	cfg = strings.TrimSpace(cfg)
	if cfg == "" {
		return
	}
	sayaCmd.Args.Append("--config", cfg)
}

func (sayaCmd *SayaCmd) WithLicenseKey(lk opaque.String) {
	lkVal := strings.TrimSpace(lk.Value())
	if lkVal == "" {
		return
	}
	sayaCmd.Args.AppendCmdArgs(&CmdArgOpaqueStr{K: "--license-key", V: opaque.NewString(lkVal)})
}

func (sayaCmd *SayaCmd) WithForgeLocation(forgeLocation string) {
	forgeLocation = strings.TrimSpace(forgeLocation)
	if forgeLocation == "" {
		return
	}
	sayaCmd.Args.Append("--forge", forgeLocation)
}

func (sayaCmd *SayaCmd) WithLogLevel(logLevel string) {
	logLevel = strings.TrimSpace(logLevel)
	if logLevel == "" {
		return
	}
	sayaCmd.Args.Append("--log-level", logLevel)
}

func (sayaCmd *SayaCmd) appendFlagIfNotBlank(key, val string) {
	val = strings.TrimSpace(val)
	if val == "" {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" {
		sayaCmd.ArgsValidationErrors = append(sayaCmd.ArgsValidationErrors,
			fmt.Errorf("sayaCmd.appendFlagIfNotBlank -- key must not be blank: key=%s", key))
		return
	}
	sayaCmd.Args.Append(key, val)
}

func (sayaCmd *SayaCmd) appendMultiFlagIfNotEmpty(key string, values []string) {
	if len(values) == 0 {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" {
		sayaCmd.ArgsValidationErrors = append(sayaCmd.ArgsValidationErrors,
			fmt.Errorf("sayaCmd.appendFlagIfNotBlank -- key must not be blank: key=%s", key))
	}

	valuesNormalized := make([]string, 0, len(values))
	for _, val := range values {
		val := strings.TrimSpace(val)
		if val != "" {
			valuesNormalized = append(valuesNormalized, val)
		}
	}

	if len(valuesNormalized) != len(values) {
		sayaCmd.ArgsValidationErrors = append(sayaCmd.ArgsValidationErrors,
			fmt.Errorf(
				"sayaCmd.appendMultiFlagIfNotEmpty -- blank value not supported: "+
					"\n\tvalues=%q \n\tnormalized=%q",
				values, valuesNormalized))
		return
	}

	sayaCmd.Args.AppendCmdArgs(&CmdMultiArgStr{K: key, V: valuesNormalized})
}

type ExecOutcome struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (sayaCmd SayaCmd) Exec(ctx context.Context) (ExecOutcome, error) {
	args := sayaCmd.Args.Args()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	execCmd := exec.CommandContext(ctx, sayaCmd.exe, args...)
	execCmd.Stderr = &stderr
	execCmd.Stdout = &stdout

	err := execCmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = -777
		var exitErr *exec.ExitError
		if found := stderrors.As(err, &exitErr); found {
			exitCode = exitErr.ExitCode()
		}
	}
	outcome := ExecOutcome{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
	if err != nil {
		wd, _ := os.Getwd()
		return outcome, errors.Wrapf(err,
			"SayaCmd.Exec -- fail to execute command:"+
				"\n\tcommand=%s \n\targ=%q \n\tpwd=%s \n\tcause=%v "+
				"\n\tstdout=%s \n\tstderr=%s",
			sayaCmd.exe, sayaCmd.Args.ArgsDisplay(), wd, err,
			stringutil.IndentN(3, string(outcome.Stdout)),
			stringutil.IndentN(3, stringutil.Truncate(outcome.Stderr, 512)))
	}
	return outcome, nil

}
