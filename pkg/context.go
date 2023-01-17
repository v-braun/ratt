package pkg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var _ types.RattContext = &rattContext{}

type rattContext struct {
	context.Context
	// Log        *Log
	eval       *hcl.EvalContext
	httpClient *http.Client
	reader     *hclparse.Parser
	id         string
	reporter   *reporter
	children   int

	bag types.ReportBag
}

func newRattContext(id string) *rattContext {
	parser := hclparse.NewParser()
	result := &rattContext{
		Context: context.Background(),
		// Log:        NewAppLog(),
		httpClient: &http.Client{},
		reader:     parser,
		reporter:   newReporter(parser),
		id:         id,
		children:   0,
		bag:        make(types.ReportBag),
		// reporter: newSimpleCliReporter(parser),
		eval: &hcl.EvalContext{
			Variables: map[string]cty.Value{
				"vars": cty.ObjectVal(map[string]cty.Value{
					"___": cty.StringVal(""),
				}),
			},
			Functions: map[string]function.Function{},
		},
	}

	return result
}

func (ctx *rattContext) NewChild(name string) types.RattContext {
	ctx.children = ctx.children + 1
	result := &rattContext{
		Context:    ctx.Context,
		eval:       ctx.eval.NewChild(),
		httpClient: ctx.httpClient,
		reader:     ctx.reader,
		id:         fmt.Sprintf("%s/%d/%s", ctx.id, ctx.children, name),
		reporter:   ctx.reporter,
		bag:        make(types.ReportBag),
	}

	if result.eval.Variables == nil {
		result.eval.Variables = map[string]cty.Value{}
	}

	return result
}

func (ctx *rattContext) RootEvalContext() *hcl.EvalContext {
	parent := ctx.eval
	for parent.Parent() != nil {
		parent = parent.Parent()
	}

	return parent
}

func (ctx *rattContext) Bag() types.ReportBag {
	return ctx.bag
}

func (ctx *rattContext) Reporter() types.Reporter {
	return ctx.reporter
}

func (ctx *rattContext) Reader() *hclparse.Parser {
	return ctx.reader
}

func (ctx *rattContext) Eval() *hcl.EvalContext {
	return ctx.eval
}

func (ctx *rattContext) HttpClient() *http.Client {
	return ctx.httpClient
}

func (ctx *rattContext) Id() string {
	return ctx.id
}
