package terraform

import (
	"context"
	"fmt"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func varFileCMDRun(cmd ...string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		varFiles := opt.Value("var-file").([]string)
		ws := opt.Value("ws").(string)

		cfg := config.ConfigFromContext(ctx)
		Logger.Printf("cfg: %s\n", cfg)

		defaultVarFiles, err := getDefaultVarFiles(cfg)
		if err != nil {
			return err
		}
		for _, v := range defaultVarFiles {
			cmd = append(cmd, "-var-file", v)
		}

		varFiles, err = VarFileIfWorkspaceSelected(cfg, ws, varFiles)
		if err != nil {
			return err
		}
		for _, v := range varFiles {
			cmd = append(cmd, "-var-file", v)
		}
		cmd = append(cmd, args...)

		ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
		wsEnv, err := getWorkspaceEnvVar(cfg, ws, varFiles)
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
}
