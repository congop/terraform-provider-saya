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

package types

import (
	"context"
	"fmt"

	"github.com/congop/terraform-provider-saya/internal/saya"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// SayaRefMustBeNameVersion implements validation for image references to require the format <name>:<version>.
type SayaRefMustBeNameVersion struct{}

func (m SayaRefMustBeNameVersion) Description(_ context.Context) string {
	return "Normalize image name to name:tag."
}

func (m SayaRefMustBeNameVersion) MarkdownDescription(_ context.Context) string {
	return "Normalize image name to name:tag."
}

func (m SayaRefMustBeNameVersion) ValidateString(
	ctx context.Context, req validator.StringRequest, resp *validator.StringResponse,
) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	refStr := req.ConfigValue.ValueString()
	switch ref, err := saya.ParseReference(refStr); {
	case err != nil:
		resp.Diagnostics.AddAttributeError(req.Path, err.Error(), fmt.Sprintf("%+v", err))
	case ref.Original != ref.Normalized():
		resp.Diagnostics.AddAttributeError(req.Path,
			"version/tag required",
			"version/version required; please use :lastest or a specific version")

	}

}
