package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func applyCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("apply", "")
	opt.SetCommandFn(applyRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func applyRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	ws := opt.Value("ws").(string)
	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg)

	if cfg.Terraform.Workspaces.Enabled {
		if !workspaceSelected() {
			if ws == "" {
				return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
			}
		}
	}

	cmd := []string{"terraform", "apply"}
	if ws == "" {
		cmd = append(cmd, "-input", ".tf.plan")
	} else {
		cmd = append(cmd, "-input", fmt.Sprintf(".tf.plan-%s", ws))
	}
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
