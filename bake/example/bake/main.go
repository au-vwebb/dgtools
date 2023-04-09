package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

var TM *dag.TaskMap

// main - This is the entry point for the application.
// For example:
//
//	$ ./bake
func Main(opt *getoptions.GetOpt) getoptions.CommandFn {
	var s string
	opt.StringVar(&s, "option", "main", opt.ValidValues("hello", "world"))
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		fmt.Printf("task: main, option: %s\n", s)
		Logger.Println(args)

		g := dag.NewGraph("greeting")
		g.TaskDependensOn(TM.Get("main:world"), TM.Get("main:hello"))
		err := g.Validate(TM)
		if err != nil {
			return fmt.Errorf("validation: %w", err)
		}
		err = g.Run(ctx, opt, args)
		if err != nil {
			return fmt.Errorf("dag err: %w", err)
		}

		return nil
	}
}

// main:world - This is a planet
func World(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		s := opt.Value("option").(string)

		fmt.Printf("task: world, option: %s\n", s)
		Logger.Println(args)

		g := dag.NewGraph("world")
		g.AddTask(TM.Get("main:hello"))
		err := g.Validate(TM)
		if err != nil {
			return fmt.Errorf("validation: %w", err)
		}
		err = g.Run(ctx, opt, args)
		if err != nil {
			return fmt.Errorf("dag err: %w", err)
		}
		return nil
	}
}
