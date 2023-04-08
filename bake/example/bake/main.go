package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/DavidGamba/dgtools/bake/lib/bake"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

var TM *dag.TaskMap

func TaskDefinitions(ctx context.Context, opt *getoptions.GetOpt) error {
	TM = dag.NewTaskMap()

	m := bake.NewTask(TM, opt, "main", "main", Main)
	bake.NewTask(TM, m, "hello", "This is a greeting", Hello)
	bake.NewTask(TM, m, "world", "This is a planet", World)

	return nil
}

func Main(opt *getoptions.GetOpt) getoptions.CommandFn {
	var s string
	opt.StringVar(&s, "option", "main", opt.ValidValues("hola", "hello"))
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		fmt.Printf("task: main, option: %s\n", s)
		Logger.Println(args)

		g := dag.NewGraph("greeting")
		g.TaskDependensOn(TM.Get("mundo"), TM.Get("hola"))
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

func Hello(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		s := opt.Value("option").(string)
		fmt.Printf("task: hola, option: %s\n", s)
		Logger.Println(args)
		return nil
	}
}

func World(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		s := opt.Value("option").(string)

		fmt.Printf("task: mundo, option: %s\n", s)
		Logger.Println(args)

		g := dag.NewGraph("mundo")
		g.AddTask(TM.Get("hola"))
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
