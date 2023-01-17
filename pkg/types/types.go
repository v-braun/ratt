package types

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

type RattContext interface {
	NewChild(name string) RattContext
	Id() string
	RootEvalContext() *hcl.EvalContext
	Reporter() Reporter
	Reader() *hclparse.Parser
	Eval() *hcl.EvalContext
	HttpClient() *http.Client
	Bag() ReportBag
}

type InvocationState = string

const (
	BeginState   InvocationState = "BeginState"
	RunningState InvocationState = "RunningState"
	EndState     InvocationState = "EndState"
	FailedState  InvocationState = "FailedState"
)

type InvocationType = string

const (
	ForEachStepInvocationType InvocationType = "foreachStep"
	ForEachInvocationType     InvocationType = "foreach"
	TestcaseInvocationType    InvocationType = "testcase"
	RequestInvocationType     InvocationType = "request"
	ScriptInvocationType      InvocationType = "script"
	AssertInvocationType      InvocationType = "assert"
	SetInvocationType         InvocationType = "set"
	// RequestInvocationType = "request"
)

type ReportBagEntry = string

const (
	IdBagEntry                   ReportBagEntry = "id"
	ItemsLengthBagEntry          ReportBagEntry = "itemsLength"
	ExpressionBagEntry           ReportBagEntry = "expression"
	DurationBagEntry             ReportBagEntry = "duration"
	SetValueBagEntry             ReportBagEntry = "setValue"
	StatusCodeBagEntry           ReportBagEntry = "statusCode"
	StatusTextBagEntry           ReportBagEntry = "statusText"
	ResponseBodyAsStringBagEntry ReportBagEntry = "responseBodyAsString"
)

type ReportBag = map[ReportBagEntry]interface{}

func GetBagEntry[T any](bag ReportBag, entry ReportBagEntry, defaultVal T) T {
	value, found := bag[entry]
	if !found {
		return defaultVal
	}

	result, ok := value.(T)
	if !ok {
		return defaultVal
	}

	return result
}

type Reporter interface {
	HasErrors() bool
	Error(msg string) Reporter
	ErrorWithRange(msg string, r *hcl.Range) Reporter

	AddDiags(diags hcl.Diagnostics) Reporter

	Invocation(iv Invocation, state InvocationState, ctx RattContext) Reporter

	GetDiagnostics() hcl.Diagnostics
	Flush(ctx RattContext) Reporter
}

type Printer interface {
	Print(log *ExecutionLog, ctx RattContext)
	Flush(log *ExecutionLog, ctx RattContext)
}

type ExecutionLogRecord struct {
	Id       string
	State    InvocationState
	Type     InvocationType
	Bag      ReportBag
	Name     string
	DefRange *hcl.Range
	Children []*ExecutionLogRecord
	// Parent   *ExecutionLogRecord
}
type ExecutionLog struct {
	Root *ExecutionLogRecord
	All  []*ExecutionLogRecord
	Head *ExecutionLogRecord
}

func NewExecutionLog() *ExecutionLog {
	result := &ExecutionLog{
		Root: nil,
		All:  []*ExecutionLogRecord{},
		Head: nil,
	}

	return result
}

func (l *ExecutionLog) Add(iv Invocation, state InvocationState, bag ReportBag) {
	// fmt.Println("BAG", bag)
	id := GetBagEntry(bag, IdBagEntry, "unknown")

	if l.Root == nil {
		l.Root = newExecutionLogRecord(iv, state, bag)
		l.All = []*ExecutionLogRecord{l.Root}
		l.Head = l.Root
		return
	}

	current, found := lo.Find(l.All, func(item *ExecutionLogRecord) bool {
		return item.Id == id
	})

	// is current
	if found {
		current.State = state
		current.Bag = bag
		l.Head = current
		return
	}

	parents := lo.Filter(l.All, func(item *ExecutionLogRecord, idx int) bool {
		return strings.HasPrefix(id, item.Id)
	})

	sort.SliceStable(parents, func(i, j int) bool {
		// longest paths comes first
		return parents[i].Id > parents[j].Id
	})

	if len(parents) > 0 {
		nearestParent := parents[0]
		newEntry := newExecutionLogRecord(iv, state, bag)
		l.Head = newEntry
		nearestParent.Children = append(nearestParent.Children, newEntry)
		l.All = append(l.All, newEntry)
	} else {
		panic(fmt.Sprintf("unexpected log entry: %s | parents: %d \n", id, len(parents)))
	}

	// is new
}

func newExecutionLogRecord(iv Invocation, state InvocationState, bag ReportBag) *ExecutionLogRecord {
	id := GetBagEntry(bag, IdBagEntry, "unknown")
	result := &ExecutionLogRecord{
		Id:       id,
		State:    state,
		Type:     iv.Type(),
		Bag:      bag,
		Name:     iv.Name(),
		DefRange: iv.DefRange(),
		Children: []*ExecutionLogRecord{},
	}

	return result
}

type Invocation interface {
	Name() string
	Type() InvocationType
	DefRange() *hcl.Range
	Exec(ctx RattContext)
}

type ExecRequestInvocation interface {
	Invocation
	ExecWithResult(ctx RattContext) *RequestExecutionResult
}
type TestCaseInvocation interface {
	Invocation
}
type ScriptInvocation interface {
	Invocation
}
type RequestInvocation interface {
	Invocation
}

type AssertInvocation interface {
	Invocation
	Expression() *hcl.Attribute
}
type SetInvocation interface {
	Invocation
	Expression() *hcl.Attribute
}

type RequestInvocationArg interface {
	Name() string
	DefRange() *hcl.Range
	ReadModel() *hcl.Attribute
}

type RequestDeclaration interface {
	GetMethod(ctx RattContext) string
	GetUrl(ctx RattContext) string
	GetContentType(ctx RattContext) (string, bool)
	GetBody(ctx RattContext) ([]byte, bool)
	GetArgs() []RequestDeclarationArg
	GetName() string
	DefRange() *hcl.Range
	GetHeaders() []RequestDeclarationHeader
}

type ScriptDeclaration interface {
	GetName() string
	DefRange() *hcl.Range
	Lang(ctx RattContext) string
	Content(ctx RattContext) string
}

type RequestDeclarationArg interface {
	Name() string
	ReadModel() *hcl.Attribute
}
type RequestDeclarationHeader interface {
	Name(ctx RattContext) string
	Value(ctx RattContext) string
	ReadModel() *hcl.BodyContent
}

type VarDeclaration interface {
	ReadModel() *hcl.Attribute
	Invocation
}

type RattModel interface {
	Vars() []VarDeclaration
	Requests() []RequestDeclaration
	Scripts() []ScriptDeclaration

	Invocations() []Invocation

	Run()
	Exec(name string, args map[string]string) *RequestExecutionResult
	MustNoErrors()

	// edit endpoints
	UpsertVar(file string, address string, value string)
}

type RequestExecutionResult struct {
	ExecTime     time.Duration
	RawHeaders   cty.Value
	Headers      cty.Value
	ResponseBody cty.Value

	Status     string
	StatusCode int

	ContentLength int64
	ContentType   string
}

type ScriptExecutionResult struct {
	ExecTime time.Duration
}
