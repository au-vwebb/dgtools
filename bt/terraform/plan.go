package terraform

import (
	"context"
	"fmt"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
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

	defaultVarFiles, err := getDefaultVarFiles(cfg)
	if err != nil {
		return err
	}
	for _, v := range defaultVarFiles {
		cmd = append(cmd, "-var-file", v)
	}

	varFiles, err = VarFileIfWorkspaceSelected(cfg, varFiles)
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
