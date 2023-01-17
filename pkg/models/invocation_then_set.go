package models

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

var _ types.AssertInvocation = &invocationThenSetModel{}

type invocationThenSetModel struct {
	readModel *hcl.Block
	exprDef   *hcl.Attribute
	name      string
}

func buildInvocationThenSetModel(block *hcl.Block, ctx types.RattContext) []types.Invocation {
	attrMap, diags := block.Body.JustAttributes()
	if ctx.Reporter().AddDiags(diags).HasErrors() {
		return nil
	}

	rootEvalCtx := ctx.RootEvalContext()
	rootVars := rootEvalCtx.Variables["vars"].AsValueMap()

	attributes := lo.Values(attrMap)
	results := lo.Map(attributes, func(attr *hcl.Attribute, i int) types.Invocation {
		_, ok := rootVars[attr.Name]
		if !ok {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("variable '%s' is not declared", attr.Name), &attr.Range)
		}
		result := &invocationThenSetModel{
			readModel: block,
			exprDef:   attr,
			name:      attr.Name,
		}

		return result
	})

	return results
}

func (set *invocationThenSetModel) Name() string {
	return set.name
}
func (set *invocationThenSetModel) Expression() *hcl.Attribute {
	return set.exprDef
}

func (set *invocationThenSetModel) Type() types.InvocationType {
	return types.SetInvocationType
}

func (set *invocationThenSetModel) DefRange() *hcl.Range {
	return &set.readModel.DefRange
}

func (set *invocationThenSetModel) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild(set.Name())

	ctx.Bag()[types.IdBagEntry] = subCtx.Id()
	ctx.Bag()[types.ExpressionBagEntry] = utils.ExpressionString(set.exprDef.Expr, ctx.Reader())

	subCtx.Reporter().Invocation(set, types.RunningState, ctx)

	rootEvalCtx := subCtx.RootEvalContext()

	rootVars := rootEvalCtx.Variables["vars"].AsValueMap()
	val, diag := set.exprDef.Expr.Value(ctx.Eval())
	if subCtx.Reporter().AddDiags(diag).HasErrors() {
		subCtx.Reporter().Invocation(set, types.FailedState, ctx)
		return
	}

	if val.Type().Equals(cty.String) {
		ctx.Bag()[types.SetValueBagEntry] = "\"" + val.AsString() + "\""
	} else if val.Type().Equals(cty.Bool) {
		ctx.Bag()[types.SetValueBagEntry] = fmt.Sprintf("%v", val.True())
	} else if val.Type().Equals(cty.Number) {
		ctx.Bag()[types.SetValueBagEntry] = fmt.Sprintf("%v", val.AsBigFloat())
	}

	rootVars[set.name] = val
	rootEvalCtx.Variables["vars"] = cty.ObjectVal(rootVars)

	subCtx.Reporter().Invocation(set, types.EndState, ctx)
}
