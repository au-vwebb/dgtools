package terraform

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func applyCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("apply", "Wrapper around terraform apply")
	opt.SetCommandFn(applyRun)

	wss := []string{}
	if cfg.Terraform.Workspaces.Enabled {
		if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
			wss, err = getWorkspaces(ctx, cfg)
			if err != nil {
				Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
				opt.String("ws", "", opt.ValidValues(wss...))
				return opt
			}
		} else {
			e, err := os.ReadFile(".terraform/environment")
			if err != nil {
				Logger.Printf("WARNING: failed to retrieve workspace: %s\n", err)
				opt.String("ws", "", opt.ValidValues(wss...))
				return opt
			}
			wss = append(wss, strings.TrimSpace(string(e)))
		}
	}
	opt.String("ws", "", opt.ValidValues(wss...))

	return opt
}

func applyRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %#v\n", cfg)

	cmd := []string{"terraform", "apply", "-input", "tf.plan"}

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
