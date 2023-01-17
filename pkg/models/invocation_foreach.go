package models

import (
	"github.com/v-braun/ratt/pkg/utils"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/v-braun/ratt/pkg/types"
)

var _ types.Invocation = &invocationForEach{}

type invocationForEach struct {
	readModel *hcl.Block
	model     types.RattModel
}

func BuildForEachInvocationModel(block *hcl.Block, model types.RattModel, ctx types.RattContext) types.Invocation {
	result := &invocationForEach{
		readModel: block,
		model:     model,
	}

	return result
}

func (ir *invocationForEach) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild("for_each")

	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	subCtx.Reporter().Invocation(ir, types.BeginState, ctx)

	body := utils.ExtractContent(ir.readModel.Body, ForEachSchema(), ctx)
	if ctx.Reporter().HasErrors() {
		subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
		return
	}

	// nested, readModel := BuildInvocationBlockBody(block, model, ForEachSchema(), ctx)
	// if ctx.Reporter().HasErrors() {
	// 	return nil
	// }

	itemsAttr, found := utils.GetOptionalAttr("items", body, subCtx)
	if ctx.Reporter().HasErrors() {
		subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
		return
	}
	if !found {
		ctx.Reporter().ErrorWithRange("attribute 'items' is missing", ir.DefRange())
		subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
		return
	}

	if !itemsAttr.Type().IsObjectType() {
		ctx.Reporter().ErrorWithRange("attribute 'items' should be an object", ir.DefRange())
		subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
		return
	}

	items := itemsAttr.AsValueMap()

	ctx.Bag()[types.ItemsLengthBagEntry] = len(items)

	for k, v := range items {
		nested, _ := BuildInvocationBlockBody(ir.readModel, ir.model, ForEachSchema(), ctx)
		if ctx.Reporter().HasErrors() {
			subCtx.Reporter().Invocation(ir, types.FailedState, ctx)
			return
		}
		step := &invocationForEachStep{
			nested:    nested,
			bodyModel: body,
			name:      k,
			values:    v,
		}

		step.Exec(subCtx)
	}

	subCtx.Reporter().Invocation(ir, types.EndState, ctx)

	// subCtx.Reporter().Invocation(ir, types.RunningState, bag)
	// start := time.Now()
	// for _, invocation := range ir.nested {
	// 	invocation.Exec(subCtx)
	// }
	// elapsed := time.Since(start)
	// bag[types.DurationBagEntry] = elapsed.Milliseconds()

	// subCtx.Reporter().Invocation(ir, types.EndState, bag)

}

func (ir *invocationForEach) Name() string {
	return "for_each"
}
func (ir *invocationForEach) Type() types.InvocationType {
	return types.ForEachInvocationType
}
func (ir *invocationForEach) DefRange() *hcl.Range {
	return &ir.readModel.DefRange
}

var _ types.Invocation = &invocationForEachStep{}

type invocationForEachStep struct {
	nested    []types.Invocation
	bodyModel *hcl.BodyContent
	name      string
	values    cty.Value
}

func (ir *invocationForEachStep) Name() string {
	return ir.name
}
func (ir *invocationForEachStep) Type() types.InvocationType {
	return types.ForEachStepInvocationType
}
func (ir *invocationForEachStep) DefRange() *hcl.Range {
	return &ir.bodyModel.MissingItemRange
}

func (ir *invocationForEachStep) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild(ir.name)

	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	subCtx.Reporter().Invocation(ir, types.BeginState, ctx)
	eval := subCtx.Eval()
	eval.Variables["each"] = cty.ObjectVal(map[string]cty.Value{
		"key":   cty.StringVal(ir.name),
		"value": ir.values,
	})

	for _, nested := range ir.nested {
		nested.Exec(subCtx)
	}

	subCtx.Reporter().Invocation(ir, types.EndState, ctx)

}
