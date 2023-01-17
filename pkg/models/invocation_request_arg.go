package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
)

type requestInvocationArg struct {
	name      string
	readModel *hcl.Attribute
	// declaration types.RequestInvocationArg
}

func BuildRequestInvocationArg(invokeContent *hcl.BodyContent, ctx types.RattContext) []types.RequestInvocationArg {
	attributes := lo.Values(invokeContent.Attributes)
	args := lo.Map(attributes, func(attr *hcl.Attribute, i int) types.RequestInvocationArg {
		result := requestInvocationArg{
			name:      attr.Name,
			readModel: attr,
			// declaration: ,
		}

		return &result
	})

	if ctx.Reporter().HasErrors() {
		return []types.RequestInvocationArg{}
	}

	return args
}

func (arg *requestInvocationArg) Name() string {
	return arg.name
}

func (arg *requestInvocationArg) ReadModel() *hcl.Attribute {
	return arg.readModel
}

func (arg *requestInvocationArg) DefRange() *hcl.Range {
	return &arg.readModel.Range
}
