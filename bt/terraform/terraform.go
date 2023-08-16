package terraform

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func NewCommand(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("terraform", "terraform related tasks")

	// backend-config
	initCMD(ctx, opt)

	// var-file
	planCMD(ctx, opt)
	importCMD(ctx, opt)

	// workspace selection
	applyCMD(ctx, opt)
	forceUnlockCMD(ctx, opt)

	return opt
}

// Retrieves workspaces assuming a convention where the .tfvars[.json] file matches the name of the workspace
// It only lists files, it doesn't query Terraform for a 'proper' list of workspaces.
func getWorkspaces(ctx context.Context, cfg *config.Config) ([]string, error) {
	wss := []string{}
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

func validWorkspaces(ctx context.Context, cfg *config.Config) ([]string, error) {
	wss := []string{}
	if cfg.Terraform.Workspaces.Enabled {
		if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
			wss, err = getWorkspaces(ctx, cfg)
			if err != nil {
				return wss, err
			}
		} else {
			e, err := os.ReadFile(".terraform/environment")
			if err != nil {
				return wss, err
			}
			wss = append(wss, strings.TrimSpace(string(e)))
		}
	}
	return wss, nil
}
