package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func forceUnlockCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("force-unlock", "")
	opt.SetCommandFn(forceUnlockRun)
	opt.HelpSynopsisArg("lock-id", "Lock ID")

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func forceUnlockRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %#v\n", cfg)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <lock-id>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	lockID := args[0]
	args = slices.Delete(args, 0, 1)

	cmd := []string{"terraform", "force-unlock", "-force", lockID}

	cmd = append(cmd, args...)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	if cfg.Terraform.Workspaces.Enabled {
		if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
			ws := opt.Value("ws").(string)
			if !opt.Called("ws") {
				return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
			}
			wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
			ri.Env(wsEnv)
			Logger.Printf("export %s\n", wsEnv)
		}
	}
	err := ri.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
