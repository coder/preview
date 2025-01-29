package preview

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform"
)

func Source(r hcl.Range, mod *terraform.Module) ([]byte, error) {
	file, ok := mod.Files[r.Filename]
	if !ok {
		return nil, os.ErrNotExist
	}

	if len(file.Bytes) < r.End.Byte {
		return nil, fmt.Errorf("range end is out of bounds")
	}

	return file.Bytes[r.Start.Byte:r.End.Byte], nil
}
