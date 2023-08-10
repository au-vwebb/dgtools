package terraform

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func initCMD(parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("init", "Wrapper around terraform init")
	opt.SetCommandFn(initRun)
	return opt
}

func initRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	f, err := config.FindFileUpwards(ctx, ".bt.cue")
	if err != nil {
		return fmt.Errorf("failed to find config file: %w", err)
	}
	Logger.Printf("Using config file: %s\n", f)
	cfg, err := config.Read(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	Logger.Printf("cfg: %#v\n", cfg)

	cmd := []string{"terraform", "init"}

	for _, bvars := range cfg.Terraform.Init.BackendConfig {
		b := strings.ReplaceAll(bvars, "~", "$HOME")
		bb, err := fsmodtime.ExpandEnv([]string{b})
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		if _, err := os.Stat(bb[0]); err == nil {
			cmd = append(cmd, "-backend-config", bb[0])
		}
	}
	cmd = append(cmd, args...)
	err = run.CMD(cmd...).Ctx(ctx).Stdin().Log().Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
