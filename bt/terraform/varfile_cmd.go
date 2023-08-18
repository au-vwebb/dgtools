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

		for _, v := range defaultVarFiles {
			cmd = append(cmd, "-var-file", v)
		}
		for _, v := range varFiles {
			cmd = append(cmd, "-var-file", v)
		}
		cmd = append(cmd, args...)

		ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
		if ws != "" {
			wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
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
