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
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/congop/terraform-provider-saya/internal/log"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

type S3FilesTransferSpecTx struct {
	Label        string
	Bucket       string
	Key          string
	WithProgress bool
	BodyProvider func() (_body io.ReadCloser, _length int64, _err error)
}
type S3FilesTransferSpec struct {
	Endpoint    *AwsEndpointSpec
	Credentials *aws.Credentials
	Region      string
	Tx          []S3FilesTransferSpecTx
}

func TransferToS3(ctx context.Context, ftSpec *S3FilesTransferSpec) error {
	for _, tx := range ftSpec.Tx {
		if tx.Bucket == "" || tx.Key == "" {
			return errors.Errorf(
				"cannot copy to s3 because bucket or key not provided: bucket=%s kex=%s",
				tx.Bucket, tx.Key)
		}

	}

	cfg, err := LoadConfig(StubLogCtx(), ftSpec.Endpoint, ftSpec.Credentials, ftSpec.Region)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	var client manager.UploadAPIClient = s3.NewFromConfig(*cfg)
	for _, tx := range ftSpec.Tx {
		err := func(tx S3FilesTransferSpecTx) error {
			body, contentLength, err := tx.BodyProvider()
			if err != nil {
				return err
			}
			defer body.Close()
			var bodySrc io.Reader = body
			if tx.WithProgress {
				bar := progressbar.NewOptions64(
					contentLength, progressbar.OptionSetDescription(tx.Label), progressbar.OptionOnCompletion(func() { fmt.Println() }))
				bodyWithProgress := progressbar.NewReader(body, bar)
				defer func() {
					if err := bar.Exit(); err != nil {
						log.Debugf(ctx, "TransferToS3 -- ignoring progress bar exit error: err=%+v", err)
					}
				}()
				bodySrc = &bodyWithProgress
			}
			input := &s3.PutObjectInput{
				Bucket:        &tx.Bucket,
				Key:           &tx.Key,
				Body:          bodySrc,
				ContentLength: contentLength,
			}

			uploader := manager.NewUploader(client, func(u *manager.Uploader) { u.Concurrency = 2 })
			_, err = uploader.Upload(StubLogCtx(), input, func(u *manager.Uploader) {
				// u.PartSize = 64 * 1024 * 1024 // 64MB per part
				u.ClientOptions = append(u.ClientOptions, func(o *s3.Options) {
					o.UsePathStyle = true
				})
			})
			if err != nil {
				return errors.Wrapf(
					err, "fail to upload to s3; bucket=%s  key=%s, kind=%s", tx.Bucket, tx.Key, tx.Label)
			}
			return nil
		}(tx)

		if err != nil {
			// TODO cleanup already done uploads?
			return err
		}

	}
	return nil
}
