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

package slices

func MapMust[T any, N any](s []T, mapper func(T) N) []N {
	if s == nil {
		return []N{}
	}
	mappedS := make([]N, 0, len(s))
	for _, e := range s {
		mapped := mapper(e)
		mappedS = append(mappedS, mapped)
	}
	return mappedS
}

func MapPMust[T any, N any](s []T, mapper func(*T) N) []N {
	if s == nil {
		return []N{}
	}
	mappedS := make([]N, 0, len(s))
	for i := 0; i < len(s); i++ {
		mapped := mapper(&s[i])
		mappedS = append(mappedS, mapped)
	}
	return mappedS
}

func FilterMust[T any](s []T, test func(T) bool) []T {
	if s == nil {
		return []T{}
	}
	mappedS := make([]T, 0, len(s))
	for _, e := range s {
		if test(e) {
			mappedS = append(mappedS, e)
		}
	}
	return mappedS
}
