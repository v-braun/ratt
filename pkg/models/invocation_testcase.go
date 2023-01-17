package models

import (
	"time"

	"github.com/samber/lo"

	"github.com/hashicorp/hcl/v2"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

var _ types.TestCaseInvocation = &invocationTestcaseModel{}

type invocationTestcaseModel struct {
	name      string
	readModel *hcl.BodyContent
	nested    []types.Invocation
}

func BuildTestcaseInvocationModels(blocks []*hcl.Block, model types.RattModel, ctx types.RattContext) []types.Invocation {
	targetBlocks := lo.Filter(blocks, func(b *hcl.Block, i int) bool {
		return b.Type == "testcase"
	})

	result := utils.MapModels(targetBlocks, ctx, func(block *hcl.Block, ctx types.RattContext) types.Invocation {
		return BuildTestcaseInvocationModel(block, model, ctx)
	})

	return result
}

func BuildTestcaseInvocationModel(block *hcl.Block, model types.RattModel, ctx types.RattContext) types.TestCaseInvocation {
	nested, readModel := BuildInvocationBlockBody(block, model, TestcaseSchema(), ctx)
	if ctx.Reporter().HasErrors() {
		return nil
	}

	result := &invocationTestcaseModel{
		readModel: readModel,
		nested:    nested,
		name:      block.Labels[0],
	}

	return result
}

func (ir *invocationTestcaseModel) Exec(ctx types.RattContext) {
	subCtx := ctx.NewChild(ir.Name())
	ctx.Bag()[types.IdBagEntry] = subCtx.Id()

	subCtx.Reporter().Invocation(ir, types.RunningState, ctx)
	start := time.Now()
	for _, invocation := range ir.nested {
		invocation.Exec(subCtx)
	}
	elapsed := time.Since(start)
	ctx.Bag()[types.DurationBagEntry] = elapsed.Milliseconds()

	subCtx.Reporter().Invocation(ir, types.EndState, ctx)

}

func (ir *invocationTestcaseModel) Name() string {
	return ir.name
}
func (ir *invocationTestcaseModel) Type() types.InvocationType {
	return types.TestcaseInvocationType
}
func (ir *invocationTestcaseModel) DefRange() *hcl.Range {
	return &ir.readModel.MissingItemRange
}
