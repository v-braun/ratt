package models

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

func BuildInvocationBlockBody(block *hcl.Block, model types.RattModel, schema *hcl.BodySchema, ctx types.RattContext) ([]types.Invocation, *hcl.BodyContent) {
	tcBody := utils.ExtractContent(block.Body, schema, ctx)
	nestedModels := make([]types.Invocation, 0)
	if ctx.Reporter().HasErrors() {
		return []types.Invocation{}, tcBody
	}

	for _, nested := range tcBody.Blocks {
		if nested.Type != "invoke" && nested.Type != "for_each" {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("block '%s' inside of testcase is not supported", nested.Type), &block.DefRange)
			continue
		}

		if nested.Type == "invoke" && len(nested.Labels) != 2 {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("block 'invoke' without type and name not supported"), &block.DefRange)
			continue
		}
		if nested.Type == "for_each" && len(nested.Labels) != 0 {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("block 'for_each' does not have any labels"), &block.DefRange)
			continue
		}

		if nested.Type == "invoke" && nested.Labels[0] == "request" {
			rqName := nested.Labels[1]
			rqDef, ok := lo.Find(model.Requests(), func(rqDef types.RequestDeclaration) bool {
				return rqDef.GetName() == rqName
			})
			if !ok {
				ctx.Reporter().ErrorWithRange(fmt.Sprintf("request declaration named '%s' could not be found", rqName), &block.DefRange)
				continue
			}

			request := BuildRequestInvocationModel(rqDef, nested, ctx)
			nestedModels = append(nestedModels, request)
			continue
		} else if nested.Type == "invoke" && nested.Labels[0] == "script" {
			scriptName := nested.Labels[1]
			scriptDef, ok := lo.Find(model.Scripts(), func(scriptDef types.ScriptDeclaration) bool {
				return scriptDef.GetName() == scriptName
			})
			if !ok {
				ctx.Reporter().ErrorWithRange(fmt.Sprintf("script declaration named '%s' could not be found", scriptDef), &block.DefRange)
				continue
			}

			request := BuildScriptInvocatioModel(scriptDef, nested, ctx)
			nestedModels = append(nestedModels, request)
			continue
		} else if nested.Type == "for_each" {
			foreachLoop := BuildForEachInvocationModel(nested, model, ctx)
			nestedModels = append(nestedModels, foreachLoop)
			continue
		} else {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("block type %s & label %s in 'testcase' not supported", block.Type, block.Labels[0]), &block.LabelRanges[0])
			continue
		}
	}

	return nestedModels, tcBody
}
