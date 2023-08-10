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
)

func planCMD(parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("plan", "Wrapper around terraform plan")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(planRun)
	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	varFiles := opt.Value("var-file").([]string)

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

	cmd := []string{"terraform", "plan", "-out", "tf.plan"}

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
	for _, v := range varFiles {
		cmd = append(cmd, "-var-file", v)
	}
	cmd = append(cmd, args...)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	if cfg.Terraform.Workspaces.Enabled {
		if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
			if len(varFiles) < 1 {
				return fmt.Errorf("running in workspace mode but no workspace selected or -var-file given")
			}
			wsFilename := filepath.Base(varFiles[0])
			r := regexp.MustCompile(`\..*$`)
			wsName := r.ReplaceAllString(wsFilename, "")
			wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", wsName)
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
