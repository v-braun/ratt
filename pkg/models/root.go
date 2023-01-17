package models

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/v-braun/ratt/pkg/types"
	"github.com/v-braun/ratt/pkg/utils"
)

func ListRootBlocks(ctx types.RattContext) []*hcl.Block {
	rootBlocks := []*hcl.Block{}
	for _, v := range ctx.Reader().Files() {
		bodyContent := utils.ExtractContent(v.Body, RootSchema(), ctx)
		if ctx.Reporter().HasErrors() {
			return rootBlocks
		}

		rootBlocks = append(rootBlocks, bodyContent.Blocks...)
	}

	return rootBlocks
}
