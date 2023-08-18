package terraform

import (
	"context"
	"fmt"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func wsCMDRun(cmd ...string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
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
}
