package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

var _ types.RequestDeclarationHeader = &requestHeaderModel{}

type requestHeaderModel struct {
	readModel *hcl.BodyContent
}

func newRequestHeaderModel(block *hcl.Block, ctx types.RattContext) *requestHeaderModel {
	content, diags := block.Body.Content(RequestHeaderSchema())
	if ctx.Reporter().AddDiags(diags).HasErrors() {
		return nil
	}

	result := &requestHeaderModel{
		readModel: content,
	}

	return result
}

func newRequestHeaderModels(blocks hcl.Blocks, ctx types.RattContext) []*requestHeaderModel {
	blocks = blocks.OfType("header")
	models := utils.MapModels(blocks, ctx, newRequestHeaderModel)
	return models
}

func (am *requestHeaderModel) Name(ctx types.RattContext) string {
	result := utils.GetRequiredAttrAsStr("name", am.readModel, ctx, utils.NoopValidate)
	return result
}

func (am *requestHeaderModel) Value(ctx types.RattContext) string {
	result := utils.GetRequiredAttrAsStr("value", am.readModel, ctx, utils.NoopValidate)
	return result
}

func (am *requestHeaderModel) ReadModel() *hcl.BodyContent {
	return am.readModel
}
