package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func planCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("plan", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(planRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	return varFileCMDRun("terraform", "plan", "-out", "tf.plan")(ctx, opt, args)
}
