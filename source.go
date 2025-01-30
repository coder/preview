package preview

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
)

func source(r hcl.Range, files map[string]*hcl.File) ([]byte, error) {
	file, ok := files[r.Filename]
	if !ok {
		return nil, os.ErrNotExist
	}

	if len(file.Bytes) < r.End.Byte {
		return nil, fmt.Errorf("range end is out of bounds")
	}

	return file.Bytes[r.Start.Byte:r.End.Byte], nil
}
