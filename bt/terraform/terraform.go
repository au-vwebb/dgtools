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

func NewCommand(parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("terraform", "terraform related tasks")

	initCMD(opt)
	planCMD(opt)
	applyCMD(opt)

	forceUnlockCMD(opt)

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
