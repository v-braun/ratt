package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
)

var _ types.RequestDeclarationArg = &requestArgModel{}

type requestArgModel struct {
	name      string
	def       *hcl.Block
	readModel *hcl.Attribute
}

func newRequestArgModels(block *hcl.Block, ctx types.RattContext) []*requestArgModel {
	attrMap, diags := block.Body.JustAttributes()
	ctx.Reporter().AddDiags(diags)
	if ctx.Reporter().HasErrors() {
		return nil
	}

	attributes := lo.Values(attrMap)

	results := lo.Map(attributes, func(attr *hcl.Attribute, i int) *requestArgModel {
		result := &requestArgModel{
			name:      attr.Name,
			def:       block,
			readModel: attr,
		}

		return result
	})

	return results
}

func (am *requestArgModel) Name() string {
	return am.name
}

func (am *requestArgModel) ReadModel() *hcl.Attribute {
	return am.readModel
}
