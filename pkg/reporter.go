package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/v-braun/ratt/pkg/types"
)

var _ types.Reporter = &reporter{}

type reporter struct {
	hcl.Diagnostics
	parser  *hclparse.Parser
	log     *types.ExecutionLog
	printer []types.Printer
}

func newReporter(parser *hclparse.Parser) *reporter {
	// printerRegion, _ := pterm.DefaultArea.Start("")
	result := &reporter{
		Diagnostics: make(hcl.Diagnostics, 0),
		parser:      parser,
		log:         types.NewExecutionLog(),
		printer:     []types.Printer{},
	}

	return result
}

func (r *reporter) addPrinter(printers ...types.Printer) types.Reporter {
	for _, p := range printers {
		r.printer = append(r.printer, p)
	}

	return r
}

func (r *reporter) Flush(ctx types.RattContext) types.Reporter {
	for _, p := range r.printer {
		p.Flush(r.log, ctx)
	}

	return r
}

func (r *reporter) Error(msg string) types.Reporter {
	r.AddDiags([]*hcl.Diagnostic{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  msg,
	}})

	return r
}

func (r *reporter) ErrorWithRange(msg string, rng *hcl.Range) types.Reporter {
	r.AddDiags([]*hcl.Diagnostic{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  msg,
		Subject:  rng,
	}})

	return r
}

func (r *reporter) AddDiags(diags hcl.Diagnostics) types.Reporter {
	r.Diagnostics = append(r.Diagnostics, diags...)
	return r
}

func (r *reporter) Invocation(iv types.Invocation, state types.InvocationState, ctx types.RattContext) types.Reporter {
	r.log.Add(iv, state, ctx.Bag())
	for _, p := range r.printer {
		p.Print(r.log, ctx)
	}

	return r
}

func (r *reporter) HasErrors() bool {
	return r.Diagnostics.HasErrors()
}

func (r *reporter) GetDiagnostics() hcl.Diagnostics {
	return r.Diagnostics
}
