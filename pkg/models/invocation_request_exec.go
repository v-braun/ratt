package models

import (
	"github.com/samber/lo"

	"github.com/hashicorp/hcl/v2"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/zclconf/go-cty/cty"
)

var _ types.ExecRequestInvocation = &invocationRequestExecModel{}

type invocationRequestExecModel struct {
	declaration types.RequestDeclaration
	args        map[string]cty.Value
	id          string
}

func BuildRequestInvocatioExecnModel(declaration types.RequestDeclaration, args map[string]string, ctx types.RattContext) types.ExecRequestInvocation {

	result := &invocationRequestExecModel{
		declaration: declaration,
		args:        map[string]cty.Value{},
	}

	result.args = lo.MapValues(args, func(v string, k string) cty.Value {
		return cty.StringVal(v)
	})

	return result
}

func (ir *invocationRequestExecModel) Exec(ctx types.RattContext) {
	ir.ExecWithResult(ctx)
}

func (ir *invocationRequestExecModel) ExecWithResult(ctx types.RattContext) *types.RequestExecutionResult {
	subCtx := ctx.NewChild(ir.Name())
	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	subCtx.Reporter().Invocation(ir, types.RunningState, ctx)

	executor := NewRequestInstance(ir.declaration)
	result := executor.Exec(subCtx, ir.args)
	if subCtx.Reporter().HasErrors() {
		subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
		return nil
	}
	ctx.Bag()[types.DurationBagEntry] = result.ExecTime.Milliseconds()

	subCtx.Reporter().Invocation(ir, types.EndState, ctx)

	subCtx.Eval().Variables["res"] = cty.ObjectVal(map[string]cty.Value{
		"statusCode":    cty.NumberIntVal(int64(result.StatusCode)),
		"status":        cty.StringVal(result.Status),
		"contentLength": cty.NumberIntVal(result.ContentLength),
		"contentType":   cty.StringVal(result.ContentType),
		"rawHeaders":    result.RawHeaders,
		"headers":       result.Headers,
		"body":          result.ResponseBody,
	})

	return result
}

func (ir *invocationRequestExecModel) Id() string {
	return ir.id
}
func (ir *invocationRequestExecModel) Name() string {
	return ir.declaration.GetName()
}
func (ir *invocationRequestExecModel) Type() types.InvocationType {
	return types.RequestInvocationType
}
func (ir *invocationRequestExecModel) DefRange() *hcl.Range {
	return ir.declaration.DefRange()
}
