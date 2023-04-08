package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"unsafe"

	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	opt := getoptions.New()
	opt.Self("bake", "Go Build + Something like Make = Bake ¯\\_(ツ)_/¯")
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))

	bakefile, err := findBakeFiles(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	plug, err := load(ctx, bakefile, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		logger, err := plug.Lookup("Logger")
		if err == nil {
			var l **log.Logger
			l, ok := logger.(*(*log.Logger))
			if ok {
				(*l).SetOutput(io.Discard)
			} else {
				Logger.Printf("failed to convert Logger: %s\n", reflect.TypeOf(logger))
			}
		} else {
			Logger.Printf("failed to find Logger\n")
		}
		Logger.SetOutput(io.Discard)
	}

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

type TaskDefinitionFn func(ctx context.Context, opt *getoptions.GetOpt) error
type TaskFn func(*getoptions.GetOpt) getoptions.CommandFn

func load(ctx context.Context, bakefile string, opt *getoptions.GetOpt) (*plugin.Plugin, error) {
	plug, err := plugin.Open(bakefile)
	if err != nil {
		return plug, fmt.Errorf("failed to open plugin: %w", err)
	}
	// inspectPlugin(plug)

	taskDefinitions, err := plug.Lookup("TaskDefinitions")
	if err != nil {
		return plug, fmt.Errorf("failed to find TaskDefinitions function: %w", err)
	}
	var tdfn TaskDefinitionFn
	tdfn, ok := taskDefinitions.(func(ctx context.Context, opt *getoptions.GetOpt) error)
	if !ok {
		return plug, fmt.Errorf("unexpected TaskDefinitions signature")
	}
	tdfn(ctx, opt)

	return plug, nil
}

// https://github.com/golang/go/issues/17823
type Plug struct {
	pluginpath string
	err        string        // set if plugin failed to load
	loaded     chan struct{} // closed when loaded
	syms       map[string]any
}

func inspectPlugin(p *plugin.Plugin) {
	pl := (*Plug)(unsafe.Pointer(p))

	Logger.Printf("Plugin %s exported symbols (%d): \n", pl.pluginpath, len(pl.syms))

	for name, pointers := range pl.syms {
		Logger.Printf("symbol: %s, pointer: %v, type: %v\n", name, pointers, reflect.TypeOf(pointers))
		if _, ok := pointers.(func(*getoptions.GetOpt) getoptions.CommandFn); ok {
			fmt.Printf("name: %s\n", name)
		}
	}
}

func findBakeFiles(ctx context.Context) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// First case, we are withing the bake folder
	base := filepath.Base(wd)
	if base == "bake" {
		err := build(".")
		if err != nil {
			return "", fmt.Errorf("failed to build: %w", err)
		}
		return "./bake.so", nil
	}

	// Second case, bake folder lives in CWD
	dir := filepath.Join(wd, "bake")
	if fi, err := os.Stat(dir); err == nil && fi.Mode().IsDir() {
		err := build(dir)
		if err != nil {
			return "", fmt.Errorf("failed to build: %w", err)
		}
		return filepath.Join(dir, "bake.so"), nil
	}

	// Third case, bake folder lives in root of repo
	root, err := buildutils.GitRepoRoot()
	if err != nil {
		return "", fmt.Errorf("failed to get git repo root: %w", err)
	}
	dir = filepath.Join(root, "bake")
	if fi, err := os.Stat(dir); err == nil && fi.Mode().IsDir() {
		err := build(dir)
		if err != nil {
			return "", fmt.Errorf("failed to build: %w", err)
		}
		return filepath.Join(dir, "bake.so"), nil
	}

	return "", fmt.Errorf("bake directory not found")
}

func build(dir string) error {
	files, modified, err := fsmodtime.Target(os.DirFS(dir),
		[]string{"bake.so"},
		[]string{"*.go", "go.mod", "go.sum"})
	if err != nil {
		return err
	}
	if modified {
		Logger.Printf("Found modifications on %v, rebuilding...\n", files)
		return run.CMD("go", "build", "-buildmode=plugin", "-o=bake.so").Dir(dir).Log().Run()
	}
	return nil
}
