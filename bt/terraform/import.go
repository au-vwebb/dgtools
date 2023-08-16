package terraform

import (
	"context"

	"github.com/DavidGamba/go-getoptions"
)

func importCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("import", "Wrapper around terraform import")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(varFileCMDRun("terraform", "import"))
	return opt
}
