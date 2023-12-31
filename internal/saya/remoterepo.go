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
	"reflect"
	"strings"
	"time"
)

const RepoTypeHttp = "http"
const RepoTypeS3 = "s3"

func IsRepoTypeHttp(repoType string) bool {
	return repoType == "http" || repoType == "https"
}

func IsRepoTypeS3(repoType string) bool {
	return repoType == "s3"
}

type AwsCredentials struct {
	// AWS Access key ID
	AccessKeyID string //revive:disable-line:var-naming

	// AWS Secret Access Key
	SecretAccessKey string

	// AWS Session Token
	SessionToken string

	// Source of the credentials
	Source string

	// Time the credentials will expire.
	CanExpire bool
	Expires   *time.Time
}

func (cred *AwsCredentials) NormalizeToNil() *AwsCredentials {
	if cred == nil {
		return nil
	}
	copy := *cred
	copy.AccessKeyID = strings.TrimSpace(copy.AccessKeyID)
	copy.SecretAccessKey = strings.TrimSpace(copy.SecretAccessKey)
	copy.SessionToken = strings.TrimSpace(copy.SessionToken)
	copy.Source = strings.TrimSpace(copy.Source)
	if copy.Expires != nil && copy.Expires.IsZero() {
		copy.Expires = nil
	}
	return &copy
}

func (cred *AwsCredentials) HasExpires() bool {
	return cred != nil && cred.Expires != nil && !cred.Expires.IsZero()
}

type AuthHttpBasic struct {
	Username string
	Pwd      string
}

type HttpRepo struct {
	RepoUrl        string
	BasePath       string
	UploadStrategy string
	AuthHttpBasic  AuthHttpBasic
}

func (repo *HttpRepo) NormalizeToNil() *HttpRepo {
	if repo == nil {
		return nil
	}
	copy := *repo
	copy.AuthHttpBasic.Pwd = strings.TrimSpace(copy.AuthHttpBasic.Pwd)
	copy.AuthHttpBasic.Username = strings.TrimSpace(copy.AuthHttpBasic.Username)
	copy.BasePath = strings.TrimSpace(copy.BasePath)
	copy.RepoUrl = strings.TrimSpace(copy.RepoUrl)
	copy.UploadStrategy = strings.TrimSpace(copy.UploadStrategy)

	if (copy == HttpRepo{}) {
		return nil
	}
	return &copy
}

type S3Repo struct {
	Bucket  string
	BaseKey string
	EpUrl   string
	EpUrlS3 string
	Region  string // aws region to send request to

	AuthAwsCreds *AwsCredentials
	UsePathStyle bool
}

func (repo *S3Repo) NormalizeToNil() *S3Repo {
	if repo == nil {
		return nil
	}
	copy := *repo
	copy.AuthAwsCreds = copy.AuthAwsCreds.NormalizeToNil()
	copy.BaseKey = strings.TrimSpace(copy.BaseKey)
	copy.Bucket = strings.TrimSpace(copy.Bucket)
	copy.EpUrl = strings.TrimSpace(copy.EpUrl)
	copy.EpUrlS3 = strings.TrimSpace(copy.EpUrlS3)
	copy.Region = strings.TrimSpace(copy.Region)

	if (copy == S3Repo{}) {
		return nil
	}
	return &copy
}

type Repos struct {
	Http *HttpRepo
	S3   *S3Repo
}

func (repo *Repos) ExactlyOneRepoSpecified() bool {
	candidates := []any{repo.Http, repo.S3}
	countSpecified := 0
	for _, r := range candidates {
		if reflect.ValueOf(r).IsNil() {
			countSpecified++
		}
	}
	return countSpecified == 1
}

// S3Only return true if only the S3 repo is specified; false otherwise.
func (repo *Repos) S3Only() bool {
	return repo != nil && repo.ExactlyOneRepoSpecified() && repo.S3 != nil
}

// HttpOnly return true if only the http repo is specified, false otherwise
func (repo *Repos) HttpOnly() bool {
	return repo != nil && repo.ExactlyOneRepoSpecified() && repo.Http != nil
}

func (repo *Repos) AvailableRepoTypes() []string {
	avails := make([]string, 0, 2)
	if repo.Http != nil {
		avails = append(avails, RepoTypeHttp)
	}

	if repo.S3 != nil {
		avails = append(avails, RepoTypeS3)
	}
	return avails
}
