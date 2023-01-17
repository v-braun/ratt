package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/models"
	"github.com/v-braun/ratt/pkg/types"
)

type rattModel struct {
	ctx *rattContext

	vars     []types.VarDeclaration
	requests []types.RequestDeclaration
	scripts  []types.ScriptDeclaration

	invocations []types.Invocation

	rootBlocks hcl.Blocks
}

var _ types.RattModel = &rattModel{}

func newModel(printer ...types.Printer) *rattModel {
	result := &rattModel{
		ctx:         newRattContext("root"),
		vars:        make([]types.VarDeclaration, 0),
		requests:    make([]types.RequestDeclaration, 0),
		scripts:     make([]types.ScriptDeclaration, 0),
		invocations: make([]types.Invocation, 0),
	}

	result.ctx.reporter.addPrinter(printer...)

	return result
}

func BuildModelFromFolder(directory string, printer ...types.Printer) types.RattModel {
	result := newModel(printer...)
	listings, err := ioutil.ReadDir(directory)
	if err != nil {
		result.ctx.reporter.Error(fmt.Sprintf("could not read directory %s / [%s]", directory, err))
		return result
	}

	files := make([]string, 0)
	for _, fi := range listings {
		if fi.IsDir() {
			continue
		}

		if filepath.Ext(fi.Name()) != ".hcl" {
			continue
		}

		path := filepath.Join(directory, fi.Name())
		files = append(files, path)
	}

	if len(files) <= 0 {
		result.ctx.reporter.Error(fmt.Sprintf("could not find any files in directory %s", directory))
		return result
	}

	buildModelFromFiles(files, result)
	return result
}

func BuildModelFromFiles(files []string, printer ...types.Printer) types.RattModel {
	result := newModel(printer...)
	buildModelFromFiles(files, result)
	return result
}

func buildModelFromFiles(files []string, m *rattModel) *rattModel {
	for _, fi := range files {
		_, diag := m.ctx.reader.ParseHCLFile(fi)
		m.ctx.reporter.AddDiags(diag)
		if m.ctx.reporter.HasErrors() {
			return m
		}
	}

	m.rootBlocks = models.ListRootBlocks(m.ctx)
	if m.ctx.reporter.HasErrors() {
		return m
	}

	m.vars = models.BuildVarModels(m.rootBlocks, m.ctx)
	if m.ctx.Reporter().HasErrors() {
		return m
	}

	m.requests = models.BuildRequestDeclarationModels(m.rootBlocks, m.ctx)
	if m.ctx.Reporter().HasErrors() {
		return m
	}

	m.scripts = models.BuildScriptDeclarationModels(m.rootBlocks, m.ctx)
	if m.ctx.Reporter().HasErrors() {
		return m
	}

	m.invocations = models.BuildTestcaseInvocationModels(m.rootBlocks, m, m.ctx)
	if m.ctx.Reporter().HasErrors() {
		return m
	}

	return m
}

func (rm *rattModel) Vars() []types.VarDeclaration {
	return rm.vars
}
func (rm *rattModel) Requests() []types.RequestDeclaration {
	return rm.requests
}
func (rm *rattModel) Scripts() []types.ScriptDeclaration {
	return rm.scripts
}
func (rm *rattModel) Invocations() []types.Invocation {
	return rm.invocations
}

func (rm *rattModel) Run() {
	for _, m := range rm.vars {
		m.Exec(rm.ctx)
	}
	for _, m := range rm.invocations {
		m.Exec(rm.ctx)
	}

	rm.ctx.Reporter().Flush(rm.ctx)
}

func (rm *rattModel) Exec(name string, args map[string]string) *types.RequestExecutionResult {
	for _, m := range rm.vars {
		m.Exec(rm.ctx)
	}

	declaration, ok := lo.Find(rm.requests, func(rq types.RequestDeclaration) bool { return rq.GetName() == name })
	if !ok {
		rm.ctx.Reporter().Error(fmt.Sprintf("Resource '%s' could not be found", name))
		return nil
	}

	rq := models.BuildRequestInvocatioExecnModel(declaration, args, rm.ctx)
	result := rq.ExecWithResult(rm.ctx)

	return result
}

func (rm *rattModel) MustNoErrors() {
	if !rm.ctx.Reporter().HasErrors() {
		return
	}

	wr := hcl.NewDiagnosticTextWriter(
		os.Stdout,               // writer to send messages to
		rm.ctx.Reader().Files(), // the parser's file cache, for source snippets
		78,                      // wrapping width
		true,                    // generate colored/highlighted output
	)
	err := wr.WriteDiagnostics(rm.ctx.Reporter().GetDiagnostics())
	if err != nil {
		log.Fatalln(err)
	} else {
		os.Exit(1)
	}
}
