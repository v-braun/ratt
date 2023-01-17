package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

var _ types.VarDeclaration = &varModel{}

type scriptModel struct {
	name string
	def  *hcl.BodyContent
}

func BuildScriptDeclarationModels(blocks []*hcl.Block, ctx types.RattContext) []types.ScriptDeclaration {
	targetBlocks := lo.Filter(blocks, func(b *hcl.Block, i int) bool {
		return b.Type == "script"
	})

	result := utils.MapModels(targetBlocks, ctx, newScriptModel)

	return result
}

func newScriptModel(block *hcl.Block, ctx types.RattContext) types.ScriptDeclaration {
	scriptBody := utils.ExtractContent(block.Body, ScriptSchema(), ctx)
	result := &scriptModel{
		name: block.Labels[0],
		def:  scriptBody,
	}

	return result
}

func (sm *scriptModel) GetName() string {
	return sm.name
}

func (sm *scriptModel) DefRange() *hcl.Range {
	return &sm.def.MissingItemRange
}

func (sm *scriptModel) Lang(ctx types.RattContext) string {
	val, has := utils.GetOptionalAttrAsString("lang", sm.def, ctx)
	if !has {
		return "js"
	}

	return val
}

func (sm *scriptModel) Content(ctx types.RattContext) string {
	val := utils.GetRequiredAttrAsStr("content", sm.def, ctx, utils.NoopValidate)
	return val
}

// Content() string
