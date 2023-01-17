package models

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type RequestInstance struct {
	declaration types.RequestDeclaration
}

func NewRequestInstance(declaration types.RequestDeclaration) *RequestInstance {
	result := &RequestInstance{
		declaration: declaration,
	}

	return result
}

func (re *RequestInstance) prepareRequest(ctx types.RattContext) *http.Request {
	urlStr := re.declaration.GetUrl(ctx)
	method := re.declaration.GetMethod(ctx)
	contentType, hasCt := re.declaration.GetContentType(ctx)
	body, hasBody := re.declaration.GetBody(ctx)
	if ctx.Reporter().HasErrors() {
		return nil
	}

	var bodyPayload io.Reader = nil
	if hasBody {
		bodyPayload = bytes.NewBuffer(body)
	}

	rq, err := http.NewRequest(method, urlStr, bodyPayload)
	if err != nil {
		ctx.Reporter().ErrorWithRange(fmt.Sprintf("failed to setup request: %s", err.Error()), re.declaration.DefRange())
		return nil
	}

	if hasCt {
		rq.Header.Add("Content-Type", contentType)
	} else if hasBody {
		rq.Header.Add("Content-Type", "application/json; charset=UTF-8")
	}

	for _, header := range re.declaration.GetHeaders() {
		name := header.Name(ctx)
		value := header.Value(ctx)
		if ctx.Reporter().HasErrors() {
			continue
		}

		rq.Header.Add(name, value)
	}

	return rq
}

func (re *RequestInstance) buildExecArgs(ctx types.RattContext, args map[string]cty.Value) *cty.Value {
	declaredArgs := re.declaration.GetArgs()
	allArgsMap := lo.KeyBy(declaredArgs, func(declaredArg types.RequestDeclarationArg) string {
		return declaredArg.Name()
	})

	argMap := lo.MapValues(allArgsMap, func(declaredArg types.RequestDeclarationArg, k string) cty.Value {
		param, ok := args[declaredArg.Name()]
		if ok {
			return param
		}

		expr := declaredArg.ReadModel().Expr
		val, diag := expr.Value(ctx.Eval())
		if ctx.Reporter().AddDiags(diag).HasErrors() {
			return cty.StringVal("")
		} else {
			return val
		}
	})

	if ctx.Reporter().HasErrors() {
		return nil
	}

	result := cty.ObjectVal(argMap)
	return &result
}

func (re *RequestInstance) Exec(ctx types.RattContext, args map[string]cty.Value) *types.RequestExecutionResult {
	execArgs := re.buildExecArgs(ctx, args)
	if ctx.Reporter().HasErrors() {
		return nil
	}

	ctx.Eval().Variables["args"] = *execArgs

	request := re.prepareRequest(ctx)
	if request == nil || ctx.Reporter().HasErrors() {
		return nil
	}

	start := time.Now()
	resp, err := ctx.HttpClient().Do(request)
	elapsed := time.Since(start)

	if err != nil {
		ctx.Reporter().ErrorWithRange(err.Error(), re.declaration.DefRange())
		return nil
	}

	headers := lo.MapValues(resp.Header, func(v []string, k string) string {
		return v[0]
	})

	ctyRawHeaders, err := utils.ConvertToCtyObject(resp.Header)
	if err != nil {
		ctx.Reporter().ErrorWithRange(fmt.Sprintf("could not store header in context %s", err), re.declaration.DefRange())
		return nil
	}

	ctyHeaders, err := utils.ConvertToCtyObject(headers)
	if err != nil {
		ctx.Reporter().ErrorWithRange(fmt.Sprintf("could not store header in context %s", err), re.declaration.DefRange())
		return nil
	}

	bodyCtxVar, err := utils.ResponseToCtyVal(resp)
	if err != nil {
		ctx.Reporter().ErrorWithRange(fmt.Sprintf("could not store response in context %s", err), re.declaration.DefRange())
		return nil
	}

	result := &types.RequestExecutionResult{
		ExecTime:     elapsed,
		Headers:      ctyHeaders,
		RawHeaders:   ctyRawHeaders,
		ResponseBody: bodyCtxVar,

		Status:     resp.Status,
		StatusCode: resp.StatusCode,

		ContentLength: resp.ContentLength,
		ContentType:   resp.Header.Get("Content-Type"),
	}

	return result
}
