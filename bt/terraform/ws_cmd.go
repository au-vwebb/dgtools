package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func wsCMDRun(cmd ...string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		cfg := config.ConfigFromContext(ctx)
		Logger.Printf("cfg: %s\n", cfg)

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
}
