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

package provider

// import (
// 	"context"

// 	"github.com/hashicorp/terraform-plugin-framework/path"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"golang.org/x/exp/maps"
// )

// func SyncAttributePlanModifier(attrPath string) planmodifier.Object {
// 	return &syncAttributePlanModifier{
// 		attrPath: attrPath,
// 	}
// }

// type syncAttributePlanModifier struct {
// 	attrPath string
// }

// func (d *syncAttributePlanModifier) Description(ctx context.Context) string {
// 	return "Ensures that attribute_one and attribute_two attributes are kept synchronized."
// }

// func (d *syncAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
// 	return d.Description(ctx)
// }

// func (d *syncAttributePlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {

// 	var attributeOne types.Object
// 	diags := req.Plan.GetAttribute(ctx, path.Root(d.attrPath), &attributeOne)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	resp.PlanValue = types.ObjectNull(maps.Clone(attributeOne.AttributeTypes(ctx))) //(map[string]attr.Type{})
// 	return

// 	// var attributeOne types.Bool
// 	// diags := req.Plan.GetAttribute(ctx, path.Root("attribute_one"), &attributeOne)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// var attributeTwo types.Bool
// 	// req.Plan.GetAttribute(ctx, path.Root("attribute_two"), &attributeTwo)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// if !attributeOne.IsNull() && !attributeTwo.IsNull() && (attributeOne.ValueBool() != attributeTwo.ValueBool()) {
// 	// 	resp.Diagnostics.AddError(
// 	// 		"attribute_one and attribute_two are both configured with different values",
// 	// 		"attribute_one is deprecated, use attribute_two instead",
// 	// 	)
// 	// 	return
// 	// }

// 	// // Default to true for both attribute_one and attribute_two when both are null.
// 	// if attributeOne.IsNull() && attributeTwo.IsNull() {
// 	// 	resp.PlanValue = types.BoolValue(true)
// 	// 	return
// 	// }

// 	// // Default to using value for attribute_two if attribute_one is null
// 	// if attributeOne.IsNull() && !attributeTwo.IsNull() {
// 	// 	resp.PlanValue = numericConfig
// 	// 	return
// 	// }

// 	// // Default to using value for attribute_one if attribute_two is null
// 	// if !attributeOne.IsNull() && attributeTwo.IsNull() {
// 	// 	resp.PlanValue = numberConfig
// 	// 	return
// 	// }
// }

// func SuppressFromDiffAttributePlanModifierStr(attrPath string) planmodifier.String {
// 	return &suppressFromDiffAttributePlanModifierStr{
// 		attrPath: attrPath,
// 	}
// }

// type suppressFromDiffAttributePlanModifierStr struct {
// 	attrPath string
// }

// func (d *suppressFromDiffAttributePlanModifierStr) Description(ctx context.Context) string {
// 	return "Ensures that attribute_one and attribute_two attributes are kept synchronised."
// }

// func (d *suppressFromDiffAttributePlanModifierStr) MarkdownDescription(ctx context.Context) string {
// 	return d.Description(ctx)
// }

// func (d *suppressFromDiffAttributePlanModifierStr) PlanModifyString(
// 	ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
// ) {

// 	var attributeOne types.String
// 	diags := req.Plan.GetAttribute(ctx, path.Root(d.attrPath), &attributeOne)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	resp.PlanValue = types.StringNull()
// 	return

// 	// var attributeOne types.Bool
// 	// diags := req.Plan.GetAttribute(ctx, path.Root("attribute_one"), &attributeOne)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// var attributeTwo types.Bool
// 	// req.Plan.GetAttribute(ctx, path.Root("attribute_two"), &attributeTwo)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// if !attributeOne.IsNull() && !attributeTwo.IsNull() && (attributeOne.ValueBool() != attributeTwo.ValueBool()) {
// 	// 	resp.Diagnostics.AddError(
// 	// 		"attribute_one and attribute_two are both configured with different values",
// 	// 		"attribute_one is deprecated, use attribute_two instead",
// 	// 	)
// 	// 	return
// 	// }

// 	// // Default to true for both attribute_one and attribute_two when both are null.
// 	// if attributeOne.IsNull() && attributeTwo.IsNull() {
// 	// 	resp.PlanValue = types.BoolValue(true)
// 	// 	return
// 	// }

// 	// // Default to using value for attribute_two if attribute_one is null
// 	// if attributeOne.IsNull() && !attributeTwo.IsNull() {
// 	// 	resp.PlanValue = numericConfig
// 	// 	return
// 	// }

// 	// // Default to using value for attribute_one if attribute_two is null
// 	// if !attributeOne.IsNull() && attributeTwo.IsNull() {
// 	// 	resp.PlanValue = numberConfig
// 	// 	return
// 	// }
// }
