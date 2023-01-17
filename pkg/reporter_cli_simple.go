package pkg

import (
	"fmt"
	"strings"

	"github.com/v-braun/ratt/pkg/types"
)

var _ types.Printer = &simpleCliPrinter{}

type simpleCliPrinter struct {
}

func NewSimpleCliPrinter() types.Printer {
	result := &simpleCliPrinter{}

	return result
}

func (r *simpleCliPrinter) Flush(log *types.ExecutionLog, ctx types.RattContext) {
}

func (r *simpleCliPrinter) Print(log *types.ExecutionLog, ctx types.RattContext) {
	r.logRecord(log.Head)
}

func (r *simpleCliPrinter) logRecord(record *types.ExecutionLogRecord) {
	if record.State == types.RunningState || record.State == types.BeginState {
		return
	}

	invocationLine := r.getInvocationLine(record)
	status := record.State

	if record.State == types.RunningState {
		status = "BEGIN"
	} else if record.State == types.EndState {
		status = "OK"
	} else if record.State == types.FailedState {
		status = "FAIL"
	}

	status = fmt.Sprintf("%-5s", status)

	output := fmt.Sprintf("%s %s", status, invocationLine)
	fmt.Println(output)
}

func (r *simpleCliPrinter) getInvocationLine(record *types.ExecutionLogRecord) string {
	typeLbl := "invoke"
	name := record.Name

	additionalLines := make([]string, 0)
	durationVal := types.GetBagEntry[int64](record.Bag, types.DurationBagEntry, -1)
	if durationVal >= 0 {
		additionalLines = append(additionalLines, fmt.Sprintf("duration: %dms", durationVal))
	}

	indentFormat := "%-10s"
	formatAdditionalLines := func() string {
		if len(additionalLines) <= 0 {
			return ""
		}

		return fmt.Sprintf("(%s)", strings.Join(additionalLines, ", "))
	}

	if record.Type == types.AssertInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "ASSERT")
		name = record.Name
		str := fmt.Sprintf("%s %s", typeLbl, name)
		return str
	} else if record.Type == types.RequestInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "REQUEST")
		name = record.Name
		status := types.GetBagEntry[string](record.Bag, types.StatusTextBagEntry, "unknown")
		additionalLines = append(additionalLines, fmt.Sprintf("status: %s", status))
		str := fmt.Sprintf("%s %s %s", typeLbl, name, formatAdditionalLines())
		return str
	} else if record.Type == types.ScriptInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "SCRIPT")
		name = record.Name
		str := fmt.Sprintf("%s %s %s", typeLbl, name, formatAdditionalLines())
		return str
	} else if record.Type == types.SetInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "SET")
		name = record.Name
		str := fmt.Sprintf("%s %s", typeLbl, name)
		return str
	} else if record.Type == types.TestcaseInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "TESTCASE")
		name = record.Name
		str := fmt.Sprintf("%s %s %s", typeLbl, name, formatAdditionalLines())
		return str
	} else if record.Type == types.ForEachInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "FOR_EACH")
		name = record.Name
		str := fmt.Sprintf("%s %s %s", typeLbl, name, formatAdditionalLines())
		return str
	} else if record.Type == types.ForEachStepInvocationType {
		typeLbl = fmt.Sprintf(indentFormat, "FOR_STEP")
		name = record.Name
		str := fmt.Sprintf("%s %s %s", typeLbl, name, formatAdditionalLines())
		return str
	} else {
		panic(fmt.Sprintf("The invocation type %s is not supported", record.Type))
	}
}
