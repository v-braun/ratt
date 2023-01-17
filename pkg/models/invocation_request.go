package models

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

var _ types.RequestInvocation = &invocationRequestModel{}

type invocationRequestModel struct {
	declaration types.RequestDeclaration
	readModel   *hcl.BodyContent
	args        []types.RequestInvocationArg
	thens       []types.Invocation
}

func BuildRequestInvocationModel(declaration types.RequestDeclaration, block *hcl.Block, ctx types.RattContext) types.RequestInvocation {
	argDefs := lo.Map(declaration.GetArgs(), func(arg types.RequestDeclarationArg, i int) hcl.AttributeSchema {
		return hcl.AttributeSchema{Name: arg.Name(), Required: false}
	})

	rqBody := utils.ExtractContent(block.Body, InvokeRequestSchema(argDefs), ctx)
	if ctx.Reporter().HasErrors() {
		return nil
	}

	result := &invocationRequestModel{
		readModel:   rqBody,
		declaration: declaration,
		args:        make([]types.RequestInvocationArg, 0),
		thens:       make([]types.Invocation, 0),
	}

	// build args
	args := BuildRequestInvocationArg(rqBody, ctx)
	if ctx.Reporter().HasErrors() {
		return nil
	}
	result.args = args

	for _, nested := range rqBody.Blocks {
		if nested.Type != "then" {
			continue
		}

		if len(nested.Labels) <= 0 {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("block 'then' without type not supported"), &block.DefRange)
			continue
		}
		if nested.Labels[0] == "assert" {
			assertInvokes := buildInvocationThenAssertModel(nested, ctx)
			result.thens = append(result.thens, assertInvokes...)
			continue
		} else if nested.Labels[0] == "set" {
			assertInvokes := buildInvocationThenSetModel(nested, ctx)
			result.thens = append(result.thens, assertInvokes...)
		} else {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("block 'then' of type %s not supported", block.Labels[0]), &block.LabelRanges[0])
			continue
		}
	}

	if ctx.Reporter().HasErrors() {
		return nil
	}

	return result
}

func (ir *invocationRequestModel) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild(ir.Name())
	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	subCtx.Reporter().Invocation(ir, types.RunningState, ctx)

	args := make(map[string]cty.Value)
	for _, arg := range ir.args {
		argVal, diag := arg.ReadModel().Expr.Value(subCtx.Eval())
		if ctx.Reporter().AddDiags(diag).HasErrors() {
			argVal = cty.StringVal("")
		}

		args[arg.Name()] = argVal
	}

	executor := NewRequestInstance(ir.declaration)
	result := executor.Exec(subCtx, args)
	if subCtx.Reporter().HasErrors() {
		subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
		return
	}
	ctx.Bag()[types.DurationBagEntry] = result.ExecTime.Milliseconds()
	ctx.Bag()[types.StatusCodeBagEntry] = result.StatusCode
	ctx.Bag()[types.StatusTextBagEntry] = result.Status
	ctx.Bag()[types.ResponseBodyAsStringBagEntry] = utils.PrettyPrintCtyVal(result.ResponseBody)

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

	for _, then := range ir.thens {
		then.Exec(subCtx)
	}

}

func (ir *invocationRequestModel) Name() string {
	return ir.declaration.GetName()
}
func (ir *invocationRequestModel) Type() types.InvocationType {
	return types.RequestInvocationType
}
func (ir *invocationRequestModel) DefRange() *hcl.Range {
	return &ir.readModel.MissingItemRange
}
