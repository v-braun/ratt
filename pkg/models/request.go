package models

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/zclconf/go-cty/cty"
	cty_json "github.com/zclconf/go-cty/cty/json"

	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/thoas/go-funk"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

var _ types.RequestDeclaration = &requestModel{}

type requestModel struct {
	name    string
	args    []*requestArgModel
	headers []*requestHeaderModel
	def     *hcl.BodyContent
}

func BuildRequestDeclarationModels(blocks []*hcl.Block, ctx types.RattContext) []types.RequestDeclaration {
	targetBlocks := lo.Filter(blocks, func(b *hcl.Block, i int) bool {
		return b.Type == "request"
	})

	result := utils.MapModels(targetBlocks, ctx, newRequestModel)

	return result
}

func newRequestModel(block *hcl.Block, ctx types.RattContext) types.RequestDeclaration {
	rqBody := utils.ExtractContent(block.Body, RequestSchema(), ctx)
	if ctx.Reporter().HasErrors() {
		return nil
	}

	result := &requestModel{
		def:  rqBody,
		args: make([]*requestArgModel, 0),
		name: block.Labels[0],
	}

	// build arg
	argBlocks := rqBody.Blocks.OfType("args")
	if ctx.Reporter().HasErrors() {
		return result
	}
	argModels := utils.MapManyModels(argBlocks, ctx, newRequestArgModels)
	result.args = argModels

	headers := newRequestHeaderModels(rqBody.Blocks, ctx)
	if ctx.Reporter().HasErrors() {
		return result
	}
	result.headers = headers

	return result
}

func (rm *requestModel) GetMethod(ctx types.RattContext) string {
	method := utils.GetRequiredAttrAsStr("method", rm.def, ctx, func(val string, body *hcl.BodyContent, attr *hcl.Attribute, ctx types.RattContext) {
		allowedMethods := []string{"POST", "PUT", "GET", "DELETE"}
		if !funk.ContainsString(allowedMethods, val) {
			ctx.Reporter().ErrorWithRange(fmt.Sprintf("method should be one of: %s got: %s", strings.Join(allowedMethods, ", "), val), &attr.Range)
		}
	})

	return method
}

func (rm *requestModel) GetUrl(ctx types.RattContext) string {
	str := utils.GetRequiredAttrAsStr("url", rm.def, ctx, func(val string, body *hcl.BodyContent, attr *hcl.Attribute, ctx types.RattContext) {
		if _, err := url.ParseRequestURI(val); err != nil {
			ctx.Reporter().ErrorWithRange(err.Error(), &attr.Range)
		}
	})

	return str
}

func (rm *requestModel) GetContentType(ctx types.RattContext) (string, bool) {
	result, has := utils.GetOptionalAttrAsString("contentType", rm.def, ctx)
	return result, has
}

func (rm *requestModel) GetBody(ctx types.RattContext) ([]byte, bool) {
	bodyVal, has := utils.GetOptionalAttr("body", rm.def, ctx)
	if !has {
		return []byte{}, false
	}

	if bodyVal.Type().Equals(cty.Bool) {
		return []byte(fmt.Sprintf("%v", bodyVal.True())), true
	} else if bodyVal.Type().Equals(cty.String) {
		return []byte(bodyVal.AsString()), true
	} else if bodyVal.Type().Equals(cty.Number) {
		return []byte(fmt.Sprintf("%v", bodyVal.AsBigFloat())), true
	}

	result, err := cty_json.Marshal(*bodyVal, bodyVal.Type())
	if err != nil {
		ctx.Reporter().ErrorWithRange(fmt.Sprintf("could not serialize body to json"), rm.DefRange())
		return []byte{}, false
	}

	return result, true
}

func (rm *requestModel) GetArgs() []types.RequestDeclarationArg {
	result := lo.Map(rm.args, func(t *requestArgModel, i int) types.RequestDeclarationArg {
		return t
	})

	return result
}

func (rm *requestModel) GetHeaders() []types.RequestDeclarationHeader {
	result := lo.Map(rm.headers, func(t *requestHeaderModel, i int) types.RequestDeclarationHeader {
		return t
	})

	return result
}

func (rm *requestModel) GetName() string {
	return rm.name
}

func (rm *requestModel) DefRange() *hcl.Range {
	return &rm.def.MissingItemRange
}
