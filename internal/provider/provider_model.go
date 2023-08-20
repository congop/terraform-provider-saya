package provider

import (
	"strings"

	"github.com/aws/smithy-go/time"
	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
)

// SayaProviderModel describes the provider data model.
// @mind no LogLevel because terraform sets it throw environment variable so we are following the lead.
type SayaProviderModel struct {
	Exe        types.String `tfsdk:"exe"`
	Config     types.String `tfsdk:"config"`
	Forge      types.String `tfsdk:"forge"`
	LicenseKey types.String `tfsdk:"license_key"`

	HttpRepo types.Object `tfsdk:"http_repo"`
	S3Repo   types.Object `tfsdk:"s3_repo"`
}

type SayaProviderModelHttpAuthBasic struct {
	Username string `tfsdk:"username"`
	Pwd      string `tfsdk:"password"`
}

type SayaProviderModelHttpRepo struct {
	RepoUrl        string                         `tfsdk:"url"`
	BasePath       string                         `tfsdk:"base_path"`
	UploadStrategy string                         `tfsdk:"upload_strategy"`
	AuthHttpBasic  SayaProviderModelHttpAuthBasic `tfsdk:"basic_auth"`
}

type SayaProviderModelS3RepoCred struct {
	AccessKeyID     string `tfsdk:"access_key_id"`
	SecretAccessKey string `tfsdk:"secret_access_key"`
	SessionToken    string `tfsdk:"session_token"`
	Source          string `tfsdk:"source"`
	CanExpire       bool   `tfsdk:"can_expire"`
	Expires         string `tfsdk:"expires"`
}

func (credTf *SayaProviderModelS3RepoCred) AsSayaCred() (*saya.AwsCredentials, error) {
	if credTf == nil {
		return nil, nil
	}
	sayaCred := saya.AwsCredentials{
		AccessKeyID:     credTf.AccessKeyID,
		SecretAccessKey: credTf.SecretAccessKey,
		SessionToken:    credTf.SessionToken,
		Source:          credTf.Source,
		CanExpire:       credTf.CanExpire,
		Expires:         nil,
	}

	if expiresTf := strings.TrimSpace(credTf.Expires); expiresTf != "" {
		expires, err := time.ParseDateTime(expiresTf)
		if err != nil {
			return nil, errors.Wrapf(err,
				"ImageResourceModelS3RepoCred.AsSayaCred -- bad date time format for expires"+
					"\n\texpected-format: RFC3339, e.g. 1985-04-12T23:20:50.52Z"+
					"\n\tvalue-string=%s \n\tparse-issue=%s",
				expiresTf, err.Error())
		}
		sayaCred.Expires = &expires
	}

	return &sayaCred, nil
}

type SayaProviderModelS3Repo struct {
	Bucket       string                       `tfsdk:"bucket"`
	BaseKey      string                       `tfsdk:"base_key"`
	EpUrlS3      string                       `tfsdk:"ep_url_s3"`
	Region       string                       `tfsdk:"region"`
	UsePathStyle bool                         `tfsdk:"use_path_style"`
	Credentials  *SayaProviderModelS3RepoCred `tfsdk:"credentials"`
}

func (repo SayaProviderModelHttpRepo) NormalizeToNil() *SayaProviderModelHttpRepo {
	repo.AuthHttpBasic.Pwd = strings.TrimSpace(repo.AuthHttpBasic.Pwd)
	repo.AuthHttpBasic.Username = strings.TrimSpace(repo.AuthHttpBasic.Username)
	repo.BasePath = strings.TrimSpace(repo.BasePath)
	repo.RepoUrl = strings.TrimSpace(repo.RepoUrl)
	repo.UploadStrategy = strings.TrimSpace(repo.UploadStrategy)

	if (repo == SayaProviderModelHttpRepo{}) {
		return nil
	}
	return &repo
}
