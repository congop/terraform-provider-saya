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

package stubrepo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/errors"
)

var (
	configLoader = config.LoadDefaultConfig
)

type AwsEndpointSpec struct {
	// default endpoint url
	Url string //revive:disable-line:var-naming
	// endpoint url to use when interaction with s3 service
	S3Url string //revive:disable-line:var-naming
	// endpoint url to use when interacting with the ec2 service
	Ec2Url string //revive:disable-line:var-naming
}

func configEndpointResolver(
	optFns []func(*config.LoadOptions) error,
	epSpec *AwsEndpointSpec,
) ([]func(*config.LoadOptions) error, error) {
	if epSpec == nil {
		return optFns, nil
	}
	optFn := func(opts *config.LoadOptions) error {
		var rfunc aws.EndpointResolverWithOptionsFunc = func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			signingName := ""
			url := epSpec.Url
			if service == "S3" {
				signingName = "S3"
				if epSpec.S3Url != "" {
					url = epSpec.S3Url
				}
			} else if service == "ec2" {
				signingName = "ec2"
				if epSpec.Ec2Url != "" {
					url = epSpec.Ec2Url
				}
			}
			if url == "" {
				err := errors.Errorf("no Url configured: epSpec=%#v", epSpec)
				return aws.Endpoint{}, &aws.EndpointNotFoundError{Err: err}
			}
			return aws.Endpoint{
				URL: url, SigningName: signingName, PartitionID: "aws",
				SigningRegion: region, Source: aws.EndpointSourceCustom,
			}, nil
		}
		opts.EndpointResolverWithOptions = rfunc
		return nil
	}
	optFns = append(optFns, optFn)
	return optFns, nil
}

func contributeCredentialsOptFn(
	optFns []func(*config.LoadOptions) error,
	credentialsSpec *aws.Credentials,
) ([]func(*config.LoadOptions) error, error) {
	if credentialsSpec == nil {

		tflog.Debug(StubLogCtx(), "contributeCredentialsOptFn -- no saya specific credential provided")
		return optFns, nil
	}
	var credFunc aws.CredentialsProviderFunc = func(context.Context) (aws.Credentials, error) {
		return *credentialsSpec, nil
	}
	opsFn := func(opts *config.LoadOptions) error {
		opts.Credentials = credFunc
		return nil
	}
	optFns = append(optFns, opsFn)
	return optFns, nil
}

func contributeRegion(
	optFns []func(*config.LoadOptions) error,
	region string,
) ([]func(*config.LoadOptions) error, error) {
	if region == "" {
		return optFns, nil
	}
	opsFn := func(opts *config.LoadOptions) error {
		opts.Region = region
		return nil
	}
	optFns = append(optFns, opsFn)
	return optFns, nil
}

func LoadConfig(
	ctx context.Context,
	endpointSpec *AwsEndpointSpec, credentialsSpec *aws.Credentials, region string,
) (cfg *aws.Config, err error) {
	optFns := make([]func(*config.LoadOptions) error, 0, 10)
	optFns, err = contributeCredentialsOptFn(optFns, credentialsSpec)
	if err != nil {
		return nil, err
	}
	optFns, err = configEndpointResolver(optFns, endpointSpec)
	if err != nil {
		return nil, err
	}

	optFns, err = contributeRegion(optFns, region)
	if err != nil {
		return nil, err
	}

	cfgVal, err := configLoader(ctx, optFns...)
	return &cfgVal, err
}

func NewS3Client(
	endpointSpec *AwsEndpointSpec, credentialsSpec *aws.Credentials, region string,
) (*s3.Client, error) {
	cfg, err := LoadConfig(StubLogCtx(), endpointSpec, credentialsSpec, region)
	if err != nil {
		return nil, errors.Wrapf(err, "NewS3Client -- fail to load aws config: err=%s", err)
	}

	client := s3.NewFromConfig(*cfg)

	return client, nil
}
