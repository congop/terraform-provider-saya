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
	"fmt"
	"net"
	"path"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
)

type ImgRegisterer interface {
	RegisterDummyImg(img *DummyImg) error
}

type RepoS3 struct {
	tc            testcontainers.Container
	portLsGateway int
	s3Client      *s3.Client
}

func (repo *RepoS3) getS3Client() (*s3.Client, error) {
	if repo.s3Client != nil {
		return repo.s3Client, nil
	}

	repos := repo.AsRepos()
	s3Repo := repos.S3

	client, err := NewS3Client(
		&AwsEndpointSpec{
			Url:   s3Repo.EpUrl,
			S3Url: s3Repo.EpUrlS3,
		},
		&awssdk.Credentials{
			AccessKeyID:     s3Repo.AuthAwsCreds.AccessKeyID,
			SecretAccessKey: s3Repo.AuthAwsCreds.SecretAccessKey,
		},
		s3Repo.Region)
	if err != nil {
		return nil, err
	}

	repo.s3Client = client
	return client, err
}

func (repo *RepoS3) mkBucket() error {
	repos := repo.AsRepos()
	s3Repo := repos.S3

	client, err := repo.getS3Client()
	if err != nil {
		return err
	}
	_, err = client.CreateBucket(StubLogCtx(),
		&s3.CreateBucketInput{
			Bucket: &s3Repo.Bucket,
		})
	if err != nil {
		return errors.Wrapf(err, "DummyRepoS3 -- fail to create bucket: bucket=%s, err=%v", s3Repo.Bucket, err)
	}

	return nil
}

func (repo *RepoS3) RegisterDummyImg(img *DummyImg) error {
	if img == nil {
		return errors.Errorf("DummyRepoS3.RegisterDummyImg -- img must nor be nil")
	}
	repos := repo.AsRepos()
	s3Repo := repos.S3

	keySegs, err := ImageRepoRelUrlSegments(
		[]string{s3Repo.BaseKey}, &img.metaData.Platform.Platform, img.metaData.Name, img.metaData.Version, img.metaData.Type)
	if err != nil {
		return err
	}
	imgKey := path.Join(keySegs...)

	txs := []S3FilesTransferSpecTx{
		{
			Label:        "uploading dummy-image",
			Bucket:       s3Repo.Bucket,
			Key:          imgKey,
			WithProgress: true,
			BodyProvider: ContentSrcBytes("image", img.img),
		},
	}
	if img.withMeta {

		metaBytes, err := img.MetaDataAsBytes()
		if err != nil {
			return err
		}
		txs = append(txs,
			S3FilesTransferSpecTx{
				Label:        "meta",
				Bucket:       s3Repo.Bucket,
				Key:          imgKey + ".meta",
				BodyProvider: ContentSrcBytes("meta", metaBytes),
			})
	}
	txSpec := &S3FilesTransferSpec{
		Endpoint: &AwsEndpointSpec{
			Url:   s3Repo.EpUrl,
			S3Url: s3Repo.EpUrlS3,
		},
		Credentials: &awssdk.Credentials{
			AccessKeyID:     s3Repo.AuthAwsCreds.AccessKeyID,
			SecretAccessKey: s3Repo.AuthAwsCreds.SecretAccessKey,
		},
		Region: s3Repo.Region,
		Tx:     txs,
	}
	if err := TransferToS3(StubLogCtx(), txSpec); err != nil {
		return err
	}

	return nil
}

func (repo *RepoS3) Close() error {
	if tc := repo.tc; tc != nil {
		err := tc.Terminate(StubLogCtx())
		if err != nil {
			return errors.Wrapf(err, "DummyRepoS3 -- fail to terminate: err =%v", err)
		}
	}
	return nil
}

func (repo *RepoS3) AsRepos() *saya.Repos {
	return &saya.Repos{
		Http: nil,
		S3: &saya.S3Repo{
			Bucket:  "repobucket",
			BaseKey: "rbase",
			EpUrl:   fmt.Sprintf("http://localhost:%d", repo.portLsGateway),
			EpUrlS3: fmt.Sprintf("http://localhost:%d/repobucket", repo.portLsGateway),
			Region:  "us-east-1",
			AuthAwsCreds: saya.AwsCredentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			},
		},
	}
}

func FreeLocalhostTcp4Port() (int, error) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return -1, errors.Wrapf(err, "FreeTcpPort -- fail to listen at a random port: %v", err)
	}
	tcpAddr := l.Addr().(*net.TCPAddr)
	port := tcpAddr.Port
	if err := l.Close(); err != nil {
		return -1, errors.Wrapf(err,
			"FreeTcpPort -- fail to close lister, port cannot be re-reused: addr=%v err=%v",
			tcpAddr, err)
	}
	return port, nil

}

func (repo *RepoS3) Start() error {
	ctx := context.Background()
	cName := "dummy-repo-s3"

	portLsGateway, err := FreeLocalhostTcp4Port()
	if err != nil {
		return err
	}
	cReq := testcontainers.ContainerRequest{
		Name:  cName,
		Image: "localstack/localstack",
		Env: map[string]string{
			"LS_LOG": "trace", "SERVICES": "dynamodb,s3,ec2",
			"DEBUG": "1", "LOCALSTACK_API_KEY": "test",
		},

		ExposedPorts: []string{
			fmt.Sprintf("%d:4566", portLsGateway), // # LocalStack Gateway port 4566
		},
	}
	sutC, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: cReq,
			Started:          false,
		})
	if err != nil {
		return errors.Wrapf(err,
			"DummyRepoS3.Start -- fail to create test container: image=%s, err(%T):%s",
			cName, err, err.Error())
	}
	repo.tc = sutC
	repo.portLsGateway = portLsGateway

	err = repo.tc.Start(StubLogCtx())
	if err != nil {
		return errors.Wrapf(err,
			"DummyRepoS3.Start -- fail to start container: req-name=%s, err(%T):%s",
			cName, err, err.Error())
	}

	if err = repo.mkBucket(); err != nil {
		return err
	}
	return nil
}
