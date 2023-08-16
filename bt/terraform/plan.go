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

func planCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("plan", "Wrapper around terraform plan")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(planRun)
	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	varFiles := opt.Value("var-file").([]string)

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %v\n", cfg)

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
	var wsEnv string
	if cfg.Terraform.Workspaces.Enabled {
		if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
			if len(varFiles) < 1 {
				return fmt.Errorf("running in workspace mode but no workspace selected or -var-file given")
			}
			wsFilename := filepath.Base(varFiles[0])
			r := regexp.MustCompile(`\..*$`)
			wsName := r.ReplaceAllString(wsFilename, "")
			wsEnv = fmt.Sprintf("TF_WORKSPACE=%s", wsName)
			Logger.Printf("export %s\n", wsEnv)
		} else {
			e, err := os.ReadFile(".terraform/environment")
			if err != nil {
				return fmt.Errorf("failed to read current workspace: %w", err)
			}
			ws := strings.TrimSpace(string(e))
			glob := fmt.Sprintf("%s/%s.tfvars*", cfg.Terraform.Workspaces.Dir, ws)
			Logger.Printf("ws: %s, glob: %s\n", ws, glob)
			ff, _, err := fsmodtime.Glob(os.DirFS("."), true, []string{glob})
			if err != nil {
				return fmt.Errorf("failed to glob ws files: %w", err)
			}
			for _, f := range ff {
				Logger.Printf("file: %s\n", f)
				if !slices.Contains(cmd, f) {
					cmd = append(cmd, "-var-file", f)
				}
			}
		}
	}
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	if wsEnv != "" {
		ri.Env(wsEnv)
	}
	err := ri.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
