package models

import (
	"github.com/hashicorp/hcl/v2"
)

func RootSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "request",
				LabelNames: []string{"name"},
			},

			{
				Type: "vars",
			},
			{
				Type:       "testcase",
				LabelNames: []string{"name"},
			},
			{
				Type:       "script",
				LabelNames: []string{"name"},
			},
		},
	}

	return schema
}

func ScriptSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "lang", Required: false},
			{Name: "content", Required: true},
		},
		// Blocks: []hcl.BlockHeaderSchema{
		// 	{
		// 		Type: "args",
		// 	},
		// },
	}

	return schema
}
func RequestSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "method", Required: true},
			{Name: "url", Required: true},
			{Name: "contentType", Required: false},
			{Name: "body", Required: false},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "args",
			},
			{
				Type: "header",
			},
		},
	}

	return schema
}

func RequestHeaderSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "name", Required: true},
			{Name: "value", Required: true},
		},
	}

	return schema
}

func InvokeRequestSchema(args []hcl.AttributeSchema) *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: args,
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "then",
				LabelNames: []string{"type"},
			},
		},
	}

	return schema
}

func InvokeScriptSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		// args,
		Blocks: []hcl.BlockHeaderSchema{
			// {
			// 	Type:       "then",
			// 	LabelNames: []string{"type"},
			// },
		},
	}

	return schema
}

func TestcaseSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "invoke",
				LabelNames: []string{"type", "name"},
			},
			{
				Type:       "for_each",
				LabelNames: []string{},
			},
		},
	}

	return schema
}

func ForEachSchema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "items",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "invoke",
				LabelNames: []string{"type", "name"},
			},
			{
				Type:       "for_each",
				LabelNames: []string{},
			},
		},
	}

	return schema
}
