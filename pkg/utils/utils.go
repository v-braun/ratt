package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/robertkrimen/otto"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

func NoopValidate(val string, body *hcl.BodyContent, attr *hcl.Attribute, ctx types.RattContext) {}

func OttoArgAsNumber(index int, defaultVal int64, call otto.FunctionCall) int64 {
	if len(call.ArgumentList) > 0 && call.ArgumentList[0].IsNumber() {
		val, err := call.ArgumentList[0].ToInteger()
		if err == nil {
			return val
		}
	}

	return defaultVal
}

func ExtractContent(body hcl.Body, schema *hcl.BodySchema, ctx types.RattContext) *hcl.BodyContent {
	content, diags := body.Content(schema)
	if ctx.Reporter().AddDiags(diags).HasErrors() {
		return content
	}

	return content
}

func GetStringAtRange(rnge hcl.Range, files map[string]*hcl.File) string {
	type contextStringer interface {
		ContextString(offset int) string
	}

	file := files[rnge.Filename]
	snipped := file.Bytes[rnge.Start.Byte:rnge.End.Byte]
	return string(snipped)

	// sc := hcl.NewRangeScanner(src, rnge.Filename, bufio.ScanLines)
}

func ExpressionString(expression hcl.Expression, parser *hclparse.Parser) string {
	exprRange := expression.Range()
	expressionStr := fmt.Sprintf("%s", GetStringAtRange(exprRange, parser.Files()))
	return expressionStr
}

func HighlightExpressionString(expression string) string {
	buf := bytes.NewBufferString("")
	quick.Highlight(buf, expression, "go", "terminal16", "hrdark")
	expression = strings.ReplaceAll(buf.String(), "\n", "")
	return expression
}

func ExtractStr(expr hcl.Expression, ctx types.RattContext) string {
	val, diag := expr.Value(ctx.Eval())
	if ctx.Reporter().AddDiags((diag)).HasErrors() {
		return ""
	}

	result := val.AsString()

	return result
}

func GetOptionalAttr(name string, body *hcl.BodyContent, ctx types.RattContext) (*cty.Value, bool) {
	attr, ok := body.Attributes[name]
	if !ok {
		return nil, false
	}

	val, diag := attr.Expr.Value(ctx.Eval())
	if ctx.Reporter().AddDiags((diag)).HasErrors() {
		return nil, false
	}

	return &val, true

}

func GetOptionalAttrAsString(name string, body *hcl.BodyContent, ctx types.RattContext) (string, bool) {
	result := ""
	defaultArg, ok := body.Attributes[name]
	if ok {
		result = ExtractStr(defaultArg.Expr, ctx)
	}

	return result, ok
}

func GetRequiredAttr(name string, attributes hcl.Attributes, r *hcl.Range, ctx types.RattContext) *hcl.Attribute {
	attr, ok := attributes[name]
	if !ok {
		ctx.Reporter().ErrorWithRange(fmt.Sprintf("%s is a required attribute", name), r)
	}

	return attr
}

func GetRequiredAttrAsStr(name string, body *hcl.BodyContent, ctx types.RattContext, validate func(val string, body *hcl.BodyContent, attr *hcl.Attribute, ctx types.RattContext)) string {
	attr := GetRequiredAttr(name, body.Attributes, &body.MissingItemRange, ctx)
	if ctx.Reporter().HasErrors() {
		return ""
	}

	str := ExtractStr(attr.Expr, ctx)
	if ctx.Reporter().HasErrors() {
		return ""
	}

	validate(str, body, attr, ctx)
	if ctx.Reporter().HasErrors() {
		return ""
	}

	return str
}

func SetVar(evalCtx *hcl.EvalContext, name string, val cty.Value) {
	m := evalCtx.Variables["vars"].AsValueMap()
	m[name] = val
	evalCtx.Variables["vars"] = cty.ObjectVal(m)
}

type MappableModel[V any] interface {
	*V // "type *T" in the current Go2 playground
}

func MapModels[R any](blocks []*hcl.Block, ctx types.RattContext, builder func(block *hcl.Block, ctx types.RattContext) R) []R {
	models := []R{}

	lo.Map(blocks, func(block *hcl.Block, i int) R {
		varM := builder(block, ctx)
		if !ctx.Reporter().HasErrors() {
			models = append(models, varM)
		}

		return varM
	})

	return models
}

func MapManyModels[R any](blocks []*hcl.Block, ctx types.RattContext, builder func(block *hcl.Block, ctx types.RattContext) []R) []R {
	models := lo.Map(blocks, func(block *hcl.Block, i int) []R {
		varM := builder(block, ctx)
		if ctx.Reporter().HasErrors() || len(varM) <= 0 {
			return nil
		}
		return varM
	})

	models = lo.Filter(models, func(v []R, i int) bool {
		return v != nil && len(v) > 0
	})

	result := lo.Flatten(models)

	return result
}

func ConvertToCtyObject(val interface{}) (cty.Value, error) {
	impliedType, err := gocty.ImpliedType(val)
	if err != nil {
		return cty.NilVal, err
	}

	result, err := gocty.ToCtyValue(val, impliedType)
	if err != nil {
		return cty.NilVal, err
	}

	return result, nil
}

func JsonToCtyObject(val []byte) (cty.Value, error) {
	impliedType, err := json.ImpliedType(val)
	if err != nil {
		return cty.NilVal, err
	}

	result, err := json.Unmarshal(val, impliedType)
	if err != nil {
		return cty.NilVal, err
	}

	return result, nil
}
func PrettyPrintCtyVal(val cty.Value) string {
	if val.Type().Equals(cty.Bool) {
		return fmt.Sprintf("%v", val.True())
	} else if val.Type().Equals(cty.String) {
		return val.AsString()
	} else if val.Type().Equals(cty.Number) {
		return fmt.Sprintf("%v", val.AsBigFloat())
	}

	result, _ := json.Marshal(val, val.Type())
	return string(result)
}

func ResponseToCtyVal(res *http.Response) (cty.Value, error) {
	contentType := res.Header.Get("Content-Type")
	result := cty.NilVal

	bodyBytes, err := io.ReadAll(res.Body)

	if err != nil {
		return result, errors.New(fmt.Sprintf("could not read response payload %s", err))
	}

	if strings.Contains(contentType, "application/json") {
		contentType = http.DetectContentType(bodyBytes)
		result, err = JsonToCtyObject(bodyBytes)
		if err != nil {
			return result, errors.New((fmt.Sprintf("could parse response as json %s", err)))
		}
	} else {
		result, err = ConvertToCtyObject(bodyBytes)
		if err != nil {
			return result, errors.New(fmt.Sprintf("response could not be converted %s", err))
		}

	}

	return result, nil
}
