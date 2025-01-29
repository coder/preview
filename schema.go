package preview

import (
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

var workspaceTagSchema = &hclext.BodySchema{
	Blocks: []hclext.BlockSchema{
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
			Body: &hclext.BodySchema{
				Mode:       hclext.SchemaJustAttributesMode,
				Attributes: []hclext.AttributeSchema{},
				//Blocks: []hclext.BlockSchema{
				//	{
				//		Type:       "tags",
				//		LabelNames: nil,
				//		Body: &hclext.BodySchema{
				//			Mode: hclext.SchemaJustAttributesMode,
				//		},
				//	},
				//},
			},
		},
	},
}

var schema = &hclext.BodySchema{
	Attributes: nil,
	Blocks: []hclext.BlockSchema{
		{
			Type: "terraform",
		},
		{
			Type: "required_providers",
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type: "locals",
		},
		{
			Type:       "output",
			LabelNames: []string{"name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "check",
			LabelNames: []string{"name"},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
			//Body:       &hclext.BodySchema{},
		},
		//{
		//	Type:       "data",
		//	LabelNames: []string{"coder_workspace_tags", "name"},
		//	Body: &hclext.BodySchema{
		//		//Mode:       hclext.SchemaJustAttributesMode,
		//		Attributes: []hclext.AttributeSchema{},
		//		Blocks: []hclext.BlockSchema{
		//			{
		//				Type:       "coder_workspace_tags",
		//				LabelNames: []string{"name"},
		//				Body: &hclext.BodySchema{
		//					Attributes: []hclext.AttributeSchema{
		//						{
		//							Name:     "tags",
		//							Required: true,
		//						},
		//					},
		//				},
		//			},
		//		},
		//	},
		//},
		{
			Type: "moved",
		},
		{
			Type: "import",
		},
		{
			Type: "removed",
		},
	},
}
