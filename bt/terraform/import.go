package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"slices"
)

func importCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("import", "Wrapper around terraform import")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(importRun)
	return opt
}

func importRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	varFiles := opt.Value("var-file").([]string)

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %#v\n", cfg)

	cmd := []string{"terraform", "import"}

	for _, vars := range cfg.Terraform.Plan.VarFile {
		v := strings.ReplaceAll(vars, "~", "$HOME")
		vv, err := fsmodtime.ExpandEnv([]string{v})
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		if _, err := os.Stat(vv[0]); err == nil {
			cmd = append(cmd, "-var-file", vv[0])
		}
	}

	varFiles, err := VarFileIfWorkspaceSelected(cfg, varFiles)
	if err != nil {
		return err
	}
	for _, v := range varFiles {
		cmd = append(cmd, "-var-file", v)
	}
	cmd = append(cmd, args...)

	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	wsEnv, err := getWorkspaceEnvVar(cfg, varFiles)
	if err != nil {
		return err
	}
	if wsEnv != "" {
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
