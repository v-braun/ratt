package models

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

var _ types.AssertInvocation = &invocationThenAssertModel{}

type invocationThenAssertModel struct {
	readModel *hcl.Block
	exprDef   *hcl.Attribute
	name      string
}

func buildInvocationThenAssertModel(block *hcl.Block, ctx types.RattContext) []types.Invocation {
	attrMap, diags := block.Body.JustAttributes()
	if ctx.Reporter().AddDiags(diags).HasErrors() {
		return nil
	}

	attributes := lo.Values(attrMap)
	results := lo.Map(attributes, func(attr *hcl.Attribute, i int) types.Invocation {
		result := &invocationThenAssertModel{
			readModel: block,
			exprDef:   attr,
			name:      attr.Name,
		}

		return result
	})

	return results
}

func (asrt *invocationThenAssertModel) Name() string {
	return asrt.name
}
func (asrt *invocationThenAssertModel) Expression() *hcl.Attribute {
	return asrt.exprDef
}

func (asrt *invocationThenAssertModel) Type() types.InvocationType {
	return types.AssertInvocationType
}

func (asrt *invocationThenAssertModel) DefRange() *hcl.Range {
	return &asrt.readModel.DefRange
}

func (asrt *invocationThenAssertModel) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild(asrt.Name())

	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	ctx.Bag()[types.ExpressionBagEntry] = utils.ExpressionString(asrt.exprDef.Expr, ctx.Reader())
	subCtx.Reporter().Invocation(asrt, types.RunningState, ctx)

	val, diag := asrt.exprDef.Expr.Value(ctx.Eval())

	if subCtx.Reporter().AddDiags(diag).HasErrors() {
		subCtx.Reporter().Invocation(asrt, types.FailedState, ctx)
		return
	}

	if val.Type() != cty.Bool {
		subCtx.Reporter().ErrorWithRange(fmt.Sprintf("expression does not return a boolean"), &asrt.exprDef.Range)
		subCtx.Reporter().Invocation(asrt, types.FailedState, ctx)
		return
	}

	if val.False() {
		subCtx.Reporter().Invocation(asrt, types.FailedState, ctx)
	} else {
		subCtx.Reporter().Invocation(asrt, types.EndState, ctx)
	}
}
