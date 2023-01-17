package pkg

import (
	"github.com/minamijoyo/hcledit/editor"
	"github.com/samber/lo"
	"github.com/v-braun/ratt/pkg/types"
)

func newDefaultClient() editor.Client {
	o := &editor.Option{}
	return editor.NewClient(o)
}

func (rm *rattModel) applyFilter(file string, filter editor.Filter) {
	c := newDefaultClient()
	err := c.Edit(file, true, filter)
	if err != nil {
		rm.ctx.Reporter().Error(err.Error())
	}
}

func (rm *rattModel) UpsertVar(file string, name string, value string) {
	_, hasExisting := lo.Find(rm.vars, func(v types.VarDeclaration) bool {
		return v.Name() == name && file == v.DefRange().Filename
	})

	address := "vars." + name
	filter := editor.NewAttributeAppendFilter(address, value, false)
	if hasExisting {
		filter = editor.NewAttributeSetFilter(address, value)
	}

	rm.applyFilter(file, filter)
}

func (rm *rattModel) RemoveVar(file string, name string) {
	_, exist := lo.Find(rm.vars, func(v types.VarDeclaration) bool {
		return v.Name() == name && file == v.DefRange().Filename
	})

	if !exist {
		return
	}

	address := "vars." + name
	filter := editor.NewAttributeRemoveFilter(address)

	rm.applyFilter(file, filter)
}
