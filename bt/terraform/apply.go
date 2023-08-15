package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func applyCMD(parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("apply", "Wrapper around terraform apply")
	opt.SetCommandFn(applyRun)

	wss := []string{}
	if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
		wss, err = getWorkspaces()
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
	opt.String("ws", "", opt.ValidValues(wss...))

	return opt
}

func getWorkspaces() ([]string, error) {
	wss := []string{}
	cfg, _, err := config.Get(context.Background(), ".bt.cue")
	if err != nil {
		return wss, fmt.Errorf("failed to find config file: %w", err)
	}
	glob := fmt.Sprintf("%s/*.tfvars*", cfg.Terraform.Workspaces.Dir)
	ff, _, err := fsmodtime.Glob(os.DirFS("."), true, []string{glob})
	if err != nil {
		return wss, fmt.Errorf("failed to glob ws files: %w", err)
	}
	for _, ws := range ff {
		ws = filepath.Base(ws)
		ws = strings.TrimSuffix(ws, ".json")
		ws = strings.TrimSuffix(ws, ".tfvars")
		wss = append(wss, ws)
	}
	return wss, nil
}

func applyRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cfg, f, err := config.Get(ctx, ".bt.cue")
	if err != nil {
		return fmt.Errorf("failed to find config file: %w", err)
	}
	Logger.Printf("Using config file: %s\n", f)
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
	err = ri.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
