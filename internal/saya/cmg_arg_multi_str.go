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

// CmdMultiArgStr models a repeatable command arg
type CmdMultiArgStr struct {
	K             string                     //Args key. e.g. --accelerated
	V             []string                   // V value. nil if arg does not allow value specification. Note that empty string "" is a valid value, and different from nil.
	ToCmdArgParts func(K, V string) []string //
	Del           bool                       // if true deleted, the arg will not be part of the final command
}

var _ ICmdArg = (*CmdMultiArgStr)(nil)

func (arg *CmdMultiArgStr) Key() string {
	return arg.K
}

func (arg *CmdMultiArgStr) ToDel() bool {
	return arg.Del
}

func (arg *CmdMultiArgStr) CmdArgs() []string {
	switch {
	case len(arg.V) == 0:
		return []string{arg.K}
	case arg.ToCmdArgParts == nil:
		args := make([]string, 0, len(arg.V)*2)
		for _, v := range arg.V {
			args = append(args, arg.K, v)
		}
		return args
	default:
		args := make([]string, 0, len(arg.V)*2)
		for _, v := range arg.V {
			args = append(args, arg.ToCmdArgParts(arg.K, v)...)
		}
		return args
	}
}

func (arg *CmdMultiArgStr) CmdArgsDisplay() []string {
	return arg.CmdArgs()
}
