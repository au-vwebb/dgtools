package terraform

import (
	"context"

	"github.com/DavidGamba/go-getoptions"
)

func planCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("plan", "Wrapper around terraform plan")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(planRun)
	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	return varFileCMDRun("terraform", "plan", "-out", "tf.plan")(ctx, opt, args)
}
