package terraform

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func planCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("plan", "")
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.StringSlice("target", 1, 99)
	opt.SetCommandFn(planRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	destroy := opt.Value("destroy").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	varFiles := opt.Value("var-file").([]string)
	targets := opt.Value("target").([]string)
	ws := opt.Value("ws").(string)
	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg)

	ws, err = getWorkspace(cfg, ws, varFiles)
	if err != nil {
		return err
	}

	defaultVarFiles, err := getDefaultVarFiles(cfg)
	if err != nil {
		return err
	}

	varFiles, err = AddVarFileIfWorkspaceSelected(cfg, ws, varFiles)
	if err != nil {
		return err
	}

	cmd := []string{"terraform", "plan"}
	for _, v := range defaultVarFiles {
		cmd = append(cmd, "-var-file", v)
	}
	for _, v := range varFiles {
		cmd = append(cmd, "-var-file", v)
	}
	if destroy {
		cmd = append(cmd, "-destroy")
	}
	if detailedExitcode {
		cmd = append(cmd, "-detailed-exitcode")
	}
	for _, t := range targets {
		cmd = append(cmd, "-target", t)
	}
	cmd = append(cmd, args...)

	if ws == "" {
		cmd = append(cmd, "-out", ".tf.plan")
	} else {
		cmd = append(cmd, "-out", fmt.Sprintf(".tf.plan-%s", ws))
	}

	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		// exit code 2 with detailed-exitcode means changes found
		var eerr *exec.ExitError
		if detailedExitcode && errors.As(err, &eerr) && eerr.ExitCode() == 2 {
			Logger.Printf("plan has changes\n")
			return eerr
		}
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
