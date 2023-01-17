package pkg

import (
	"fmt"
	"strings"

	tm "github.com/buger/goterm"
	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

var _ types.Printer = &liveCliPrinter{}

type liveCliPrinter struct {
	area *pterm.AreaPrinter
}

func NewLiveCliPrinter() types.Printer {
	printerRegion, _ := pterm.DefaultArea.Start("")

	result := &liveCliPrinter{
		area: printerRegion,
	}

	return result
}

func (r *liveCliPrinter) Flush(log *types.ExecutionLog, ctx types.RattContext) {
}

func (r *liveCliPrinter) Print(log *types.ExecutionLog, ctx types.RattContext) {
	output := r.printRecursive(log.Root, 0)
	r.area.Update(output)
}

func (r *liveCliPrinter) printRecursive(record *types.ExecutionLogRecord, depth int) string {

	invocationLines := r.getInvocationLines(record)

	ico := "⦿"

	if record.State == types.RunningState {
		ico = color.New(color.FgHiCyan).Sprintf("⦿")
	} else if record.State == types.EndState {
		ico = color.New(color.FgGreen).Sprintf("✔")
	} else if record.State == types.FailedState {
		ico = color.New(color.FgRed).Sprintf("✖")
	}

	padding := strings.Repeat(" ", depth*2)
	output := ""
	for i, line := range invocationLines {
		if i == 0 { // first line
			output = fmt.Sprintf("%s %s %s\n", padding, ico, line)
		} else { // all others
			output = output + fmt.Sprintf("%s    %s\n", padding, line)
		}
	}

	// id := types.GetBagEntry(record.Bag, types.IdBagEntry, "NOTSET")
	// tm.Println(output)

	for _, sub := range record.Children {
		res := r.printRecursive(sub, depth+1)
		output += res
	}

	return output
}

func (r *liveCliPrinter) getInvocationLines(record *types.ExecutionLogRecord) []string {
	styleMuted := color.New(color.FgHiBlack).Sprintf
	styleComment := color.New(color.FgHiBlack, color.Italic).Sprintf

	typeLbl := styleMuted("invoke")
	name := color.New(color.FgMagenta, color.Bold).Sprintf(record.Name)
	expression := types.GetBagEntry[string](record.Bag, types.ExpressionBagEntry, "")

	durationVal := types.GetBagEntry[int64](record.Bag, types.DurationBagEntry, -1)
	duration := "unknown"
	if durationVal >= 0 {
		duration = fmt.Sprintf("%dms", durationVal)
	}
	// duration = color.New(color.FgHiBlue, color.Italic).Sprintf(duration)

	// location := styleComment("// in %s line %d", iv.DefRange().Filename, iv.DefRange().Start.Line)

	if record.Type == types.AssertInvocationType {
		typeLbl = color.New(color.FgHiBlack).Sprint("assert")
		name = color.New(color.FgHiYellow).Sprintf(record.Name)
		expression = utils.HighlightExpressionString(expression)
		str := fmt.Sprintf("%s %s %s", typeLbl, name, expression)
		return []string{str}
	} else if record.Type == types.RequestInvocationType {
		status := types.GetBagEntry[string](record.Bag, types.StatusTextBagEntry, "unknown")

		typeLbl = color.New(color.FgHiBlack).Sprint("invoke request")
		name = color.New(color.FgHiCyan).Sprintf(record.Name)

		hintLine := color.New(color.FgHiBlue, color.Italic).
			Sprintf("(duration: %s, status: %s)", duration, status)

		mainLine := fmt.Sprintf("%s %s %s", typeLbl, name, hintLine)
		respBody := types.GetBagEntry[string](record.Bag, types.ResponseBodyAsStringBagEntry, "")
		if respBody == "" {
			return []string{mainLine}
		} else {
			maxLen := int(float32(tm.Width()) / 1.5)
			if len(respBody) > maxLen {
				respBody = respBody[0:maxLen] + " ... (more)"
			}

			// return []string{mainLine}
			respBody = styleComment("=> %s", respBody)
			return []string{mainLine, respBody}
		}

	} else if record.Type == types.ScriptInvocationType {
		typeLbl = color.New(color.FgHiBlack).Sprint("invoke script")
		name = color.New(color.FgHiRed).Sprintf(record.Name)
		str := fmt.Sprintf("%s %s %s", typeLbl, name, duration)
		return []string{str}
	} else if record.Type == types.SetInvocationType {
		typeLbl = color.New(color.FgHiBlack).Sprint("set")
		name = color.New(color.FgHiGreen).Sprintf(record.Name)
		expression = utils.HighlightExpressionString(expression)
		str := fmt.Sprintf("%s %s = %s", typeLbl, name, expression)
		return []string{str}
	} else if record.Type == types.ForEachInvocationType {
		typeLbl = color.New(color.FgHiBlack).Sprint("foreach")
		length := types.GetBagEntry[int](record.Bag, types.ItemsLengthBagEntry, -1)
		lengthLbl := color.New(color.FgHiBlue, color.Italic).
			Sprintf("(%d items)", length)

		str := fmt.Sprintf("%s %s", typeLbl, lengthLbl)
		return []string{str}
	} else if record.Type == types.ForEachStepInvocationType {
		name = color.New(color.FgHiBlack).Sprint(record.Name)
		str := fmt.Sprintf("%s", name)
		return []string{str}
	} else if record.Type == types.TestcaseInvocationType {
		typeLbl = color.New(color.FgHiBlack).Sprint("invoke testcase")
		name = color.New(color.FgHiMagenta).Sprintf(record.Name)
		str := fmt.Sprintf("%s %s %s", typeLbl, name, duration)
		return []string{str}
	} else {
		panic(fmt.Sprintf("The invocation type %s is not supported", record.Type))
	}

}
