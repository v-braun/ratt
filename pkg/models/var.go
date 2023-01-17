package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"

	"github.com/zclconf/go-cty/cty"
)

var _ types.VarDeclaration = &varModel{}

type varModel struct {
	name string
	def  *hcl.Attribute
}

func newVarModel(block *hcl.Block, ctx types.RattContext) []types.VarDeclaration {
	result := make([]types.VarDeclaration, 0)
	attributes, diags := block.Body.JustAttributes()
	if ctx.Reporter().AddDiags(diags).HasErrors() {
		return nil
	}

	for k, v := range attributes {
		entry := &varModel{
			name: k,
			def:  v,
		}

		utils.SetVar(ctx.Eval(), entry.name, cty.NilVal)
		result = append(result, entry)
	}

	return result
}

func BuildVarModels(blocks hcl.Blocks, ctx types.RattContext) []types.VarDeclaration {
	blocks = blocks.OfType("vars")
	models := utils.MapManyModels(blocks, ctx, newVarModel)
	modelsByName := lo.GroupBy(models, func(t types.VarDeclaration) string { return t.Name() })
	for _, v := range modelsByName {
		if len(v) > 1 {
			ctx.Reporter().ErrorWithRange("duplicate variable declaration", &v[1].ReadModel().NameRange)
		}
	}

	return models
}

func (vm *varModel) Exec(ctx types.RattContext) {
	evalCtx := ctx.Eval()
	if evalCtx.Variables == nil {
		evalCtx.Variables = map[string]cty.Value{}
	}

	val, diag := vm.def.Expr.Value(evalCtx)
	if ctx.Reporter().AddDiags(diag).HasErrors() {
		return
	}

	utils.SetVar(evalCtx, vm.name, val)
}

func (vm *varModel) DefRange() *hcl.Range {
	return &vm.def.Range
}

func (vm *varModel) Type() string {
	return "var"
}

func (vm *varModel) Name() string {
	return vm.name
}

func (vm *varModel) ReadModel() *hcl.Attribute {
	return vm.def
}
