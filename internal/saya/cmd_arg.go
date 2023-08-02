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

// taken from https://github.com/congop/go-virtualbox/blob/master/cmd_arg.go
package saya

import "github.com/congop/terraform-provider-saya/internal/opaque"

type ICmdArg interface {
	Key() string
	ToDel() bool
	CmdArgs() []string
	CmdArgsDisplay() []string
}

// CmdArgStr models a command arg, which can be a flag or not.
type CmdArgStr struct {
	K             string                     //Args key. e.g. --accelerated
	V             *string                    // V value. nil if arg does not allow value specification. Note that empty string "" is a valid value, and different from nil.
	ToCmdArgParts func(K, V string) []string //
	Del           bool                       // if true deleted, the arg will not be part of the final command
}

func (arg *CmdArgStr) Key() string {
	return arg.K
}

func (arg *CmdArgStr) ToDel() bool {
	return arg.Del
}

func (arg *CmdArgStr) CmdArgs() []string {
	switch {
	case arg.V == nil:
		return []string{arg.K}
	case arg.ToCmdArgParts == nil:
		return []string{arg.K, *arg.V}
	default:
		return arg.ToCmdArgParts(arg.K, *arg.V)
	}
}

func (arg *CmdArgStr) CmdArgsDisplay() []string {
	return arg.CmdArgs()
}

type CmdArgOpaqueStr struct {
	K             string                                   //Args key. e.g. --accelerated
	V             *opaque.String                           // V value. nil if arg does not allow value specification. Note that empty string "" is a valid value, and different from nil.
	ToCmdArgParts func(K string, V opaque.String) []string //
	Del           bool                                     // if true deleted, the arg will not be part of the final command
}

func (arg *CmdArgOpaqueStr) Key() string {
	return arg.K
}

func (arg *CmdArgOpaqueStr) ToDel() bool {
	return arg.Del
}

func (arg *CmdArgOpaqueStr) CmdArgs() []string {
	switch {
	case arg.V == nil:
		return []string{arg.K}
	case arg.ToCmdArgParts == nil:
		return []string{arg.K, arg.V.Value()}
	default:
		return arg.ToCmdArgParts(arg.K, *arg.V)
	}
}

func (arg *CmdArgOpaqueStr) CmdArgsDisplay() []string {
	switch {
	case arg.V == nil:
		return []string{arg.K}
	case arg.ToCmdArgParts == nil:
		return []string{arg.K, arg.V.String()}
	default:
		return arg.ToCmdArgParts(arg.K, *arg.V)
	}
}

type CmdArgs struct {
	args      []ICmdArg
	overrides []ICmdArg
}

func NewCmdArgNoValue(k string) *CmdArgStr {
	return &CmdArgStr{K: k}
}

func NewCmdArg(k, v string) *CmdArgStr {
	return &CmdArgStr{K: k, V: &v}
}

func NewCmdArgDeleted(k string) CmdArgStr {
	return CmdArgStr{K: k, V: nil, Del: true}
}

func (cmdArgs *CmdArgs) AppendNoValue(key string) {
	cmdArgs.args = append(cmdArgs.args, NewCmdArgNoValue(key))
}

func (cmdArgs *CmdArgs) Append(key, value string) {
	cmdArgs.args = append(cmdArgs.args, NewCmdArg(key, value))
}

func (cmdArgs *CmdArgs) AppendCmdArgs(arg ...ICmdArg) {
	if len(arg) == 0 {
		return
	}
	cmdArgs.args = append(cmdArgs.args, arg...)
}

func (cmdArgs *CmdArgs) AppendOverride(arg ...ICmdArg) {
	if len(arg) == 0 {
		return
	}
	cmdArgs.overrides = append(cmdArgs.overrides, arg...)
}

// Args returns an slice containing the args which can be use in a command execution context.
// Args with multiple occurrence are not supported we will be overriding values.
func (cmdArgs CmdArgs) Args() []string {
	return cmdArgs.argsByFn(ICmdArg.CmdArgs)
}

func (cmdArgs CmdArgs) ArgsDisplay() []string {
	return cmdArgs.argsByFn(ICmdArg.CmdArgsDisplay)
}

func (cmdArgs CmdArgs) argsByFn(toCmdArgsFn func(ICmdArg) []string) []string {
	m := make(map[string]ICmdArg, len(cmdArgs.args)+len(cmdArgs.overrides))
	orderK := make([]string, 0, len(cmdArgs.args)+len(cmdArgs.overrides))
	for _, curArgs := range [][]ICmdArg{cmdArgs.args, cmdArgs.overrides} {
		for _, arg := range curArgs {
			k := arg.Key()
			_, contains := m[k]
			if !contains {
				orderK = append(orderK, k)
			}
			// yes we are always overriding --> multiple occurrences are not supported
			m[k] = arg
		}
	}
	argStrs := make([]string, 0, len(orderK))
	for _, k := range orderK {
		arg := m[k]
		if arg.ToDel() {
			continue
		}
		argStrs = append(argStrs, arg.CmdArgs()...)
	}
	return argStrs
}
