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
	"bytes"
	"io"

	"github.com/congop/terraform-provider-saya/internal/random"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/exp/slices"
)

type RemoteRepos struct {
	Http *HttpRepo
	S3   *RepoS3
}

func (drs *RemoteRepos) Close() error {
	errs := make([]error, 0, 2)
	for _, c := range []io.Closer{drs.Http, drs.S3} {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return multierror.Append(nil, errs...)
	}
}

func (drs *RemoteRepos) GivenHttpRepoStarted() error {
	r := NewDummyHttpRepo()
	if err := r.Start(); err != nil {
		return err
	}

	drs.Http = r
	return nil
}

func (drs *RemoteRepos) GivenS3RepoStarted() error {
	r := &RepoS3{}
	if err := r.Start(); err != nil {
		return err
	}

	drs.S3 = r
	return nil
}

type ImgInRepo struct {
	img    DummyImg
	Sha256 string
	Repos  *saya.Repos
}

func (r *ImgInRepo) HasHttpRepo() bool {
	return r != nil && r.Repos != nil && r.Repos.Http != nil
}

func GivenImgInRemoteRepo(
	imgRegisterer func(*DummyImg) error,
	specData PullImgSpecData, metaInRemoteRepo bool,
) (*ImgInRepo, error) {
	var remoteImg *DummyImg
	var imgBytes []byte
	randStr, err := random.String(16)
	if err != nil {
		return nil, err
	}
	imgBytes = []byte(specData.Tag.Normalized() + randStr)
	remoteImg, err = NewDummyImg(imgBytes, metaInRemoteRepo, specData, nil)
	if err != nil {
		return nil, err
	}

	if err = imgRegisterer(remoteImg); err != nil {
		return nil, err
	}
	sha256, err := DigestSha256(bytes.NewBuffer(slices.Clone(imgBytes)))
	if err != nil {
		return nil, err
	}
	ret := &ImgInRepo{
		img:    *remoteImg,
		Sha256: sha256,
	}
	return ret, nil
}
