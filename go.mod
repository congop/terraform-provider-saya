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

module github.com/congop/terraform-provider-saya

go 1.19

require (
	github.com/aws/aws-sdk-go-v2 v1.21.2
	github.com/aws/aws-sdk-go-v2/config v1.19.0
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.91
	github.com/aws/aws-sdk-go-v2/service/s3 v1.40.2
	github.com/aws/smithy-go v1.15.0
	github.com/containerd/containerd v1.7.3
	github.com/hashicorp/terraform-plugin-docs v0.16.0
	github.com/hashicorp/terraform-plugin-framework v1.3.5
	github.com/hashicorp/terraform-plugin-go v0.18.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-testing v1.4.0
	github.com/opencontainers/image-spec v1.1.0-rc4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.45 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.1.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.38 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.15.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.23.2 // indirect
	github.com/bytedance/sonic v1.10.0-rc3 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.0 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.5+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.1.8 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/term v0.10.0 // indirect
	golang.org/x/tools v0.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230726155614-23370e0ffb3e // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230717121422-5aa5874ade95 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/congop/execstub v0.0.0-20210519151710-e93c592abc71
	github.com/fatih/color v1.15.0 // indirect
	github.com/gin-gonic/gin v1.9.1
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320 // indirect
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.10 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hc-install v0.5.2 // indirect
	github.com/hashicorp/hcl/v2 v2.17.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.18.1 // indirect
	github.com/hashicorp/terraform-json v0.17.1 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.27.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.1 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jarcoal/httpmock v1.3.1
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/cli v1.1.5 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/posener/complete v1.2.3 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/stretchr/testify v1.8.4
	github.com/testcontainers/testcontainers-go v0.23.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/zclconf/go-cty v1.13.2 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.13.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	k8s.io/apimachinery v0.28.0
)
