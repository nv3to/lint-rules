package imports

import (
	"fmt"
	"os"

	"example.com/mod/pkg"
	"example.com/mod/pkg/sub"
	"example.com/other/lib"
)

var _ = []any{fmt.Sprint, os.Exit, pkg.X, sub.X, lib.X}
