package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/robertkrimen/otto"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

type ScriptInstance struct {
	declaration types.ScriptDeclaration
}

func NewScriptInstance(declaration types.ScriptDeclaration) *ScriptInstance {
	result := &ScriptInstance{
		declaration: declaration,
	}

	return result
}

func (si *ScriptInstance) registerDefaultFunctions(vm *otto.Otto) error {
	err := vm.Set("sleep", func(call otto.FunctionCall) otto.Value {
		ms := utils.OttoArgAsNumber(0, 0, call)
		time.Sleep(time.Millisecond * time.Duration(ms))
		return otto.Value{}
	})
	if err != nil {
		return errors.New(fmt.Sprintf("failed register 'sleep' func: %s", err.Error()))
	}
	err = vm.Set("__ratt__console_log", func(call otto.FunctionCall) otto.Value {
		printArgs := lo.Map(call.ArgumentList, func(val otto.Value, idx int) interface{} {
			str, err := val.ToString()
			if err != nil {
				return err
			} else {
				return str
			}
		})

		fmt.Println(printArgs)
		return otto.Value{}
	})
	if err != nil {
		return errors.New(fmt.Sprintf("failed register 'console log' func: %s", err.Error()))
	}

	return nil
}

func (si *ScriptInstance) registerGlobalObjects(vm *otto.Otto) error {
	registerScript := `
console = {log: __ratt__console_log};
`

	_, err := vm.Run(registerScript)
	if err != nil {
		return errors.New(fmt.Sprintf("failed register 'globals script': %s", err.Error()))
	}

	return nil
}

func (si *ScriptInstance) buildEcmaScriptVm() (*otto.Otto, error) {
	vm := otto.New()
	err := si.registerDefaultFunctions(vm)
	if err != nil {
		return nil, err
	}

	err = si.registerGlobalObjects(vm)
	if err != nil {
		return nil, err
	}

	return vm, err

	// vm.Set("___invoke_myTestRequest", func(call otto.FunctionCall) otto.Value {
	// 	fmt.Println("call invoke", call.ArgumentList[0])
	// 	time.Sleep(10 * time.Second)
	// 	return otto.Value{}
	// })

	// requests := map[string]interface{}{
	// 	"myTestRequest": map[string]interface{}{
	// 		"invoke": "___invoke_myTestRequest()",
	// 	},
	// }
	// 	vm.Run(`
	// requests = {
	// 	myTestRequest: {
	// 		invoke: ___invoke_myTestRequest
	// 	}
	// }
	// `)

}

func (si *ScriptInstance) runScript(ctx types.RattContext, content string) (*types.ScriptExecutionResult, error) {
	result := &types.ScriptExecutionResult{}

	// time.Sleep(time.Second * 2)
	// fmt.Println("run", content)
	// time.Sleep(time.Second * 2)
	vm, err := si.buildEcmaScriptVm()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("script '%s' failed: %s", si.declaration.GetName(), err.Error()))
	}

	_, err = vm.Run(content)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("script '%s' failed: %s", si.declaration.GetName(), err.Error()))
	}

	return result, nil
}

func (si *ScriptInstance) Exec(ctx types.RattContext) (*types.ScriptExecutionResult, error) {

	start := time.Now()
	result, err := si.runScript(ctx, si.declaration.Content(ctx))
	elapsed := time.Since(start)

	if err != nil {
		return nil, err
	}

	result.ExecTime = elapsed

	return result, nil
}
