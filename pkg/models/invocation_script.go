package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

var _ types.ScriptInvocation = &invocationScriptModel{}

type invocationScriptModel struct {
	declaration types.ScriptDeclaration
	readModel   *hcl.BodyContent
}

func BuildScriptInvocatioModel(declaration types.ScriptDeclaration, block *hcl.Block, ctx types.RattContext) types.ScriptInvocation {
	scriptBody := utils.ExtractContent(block.Body, InvokeScriptSchema(), ctx)
	result := &invocationScriptModel{
		declaration: declaration,
		readModel:   scriptBody,
	}

	return result
}

func (is *invocationScriptModel) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild(is.Name())
	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	subCtx.Reporter().Invocation(is, types.RunningState, ctx)

	executor := NewScriptInstance(is.declaration)
	result, err := executor.Exec(subCtx)
	if err != nil {
		subCtx.Reporter().Invocation(is, types.FailedState, ctx)
		subCtx.Reporter().ErrorWithRange(err.Error(), is.DefRange())
		return
	}
	ctx.Bag()[types.DurationBagEntry] = result.ExecTime.Milliseconds()

	subCtx.Reporter().Invocation(is, types.EndState, ctx)
}

func (is *invocationScriptModel) Name() string {
	return is.declaration.GetName()
}
func (is *invocationScriptModel) Type() types.InvocationType {
	return types.ScriptInvocationType
}
func (is *invocationScriptModel) DefRange() *hcl.Range {
	return is.declaration.DefRange()
}
