package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func importCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("import", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(importRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

type importFns struct{}

func (fn importFns) cmdFunction(ws string) []string {
	return []string{}
}

func (fn importFns) errorFunction(ws string) {
	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}
	os.Remove(planFile)
}

func (fn importFns) successFunction(ws string) {
	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}
	os.Remove(planFile)
}

func importRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	i := importFns{}
	return varFileCMDRun(i, "terraform", "import")(ctx, opt, args)
}
