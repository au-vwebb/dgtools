package bake

import (
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

type TaskFn func(*getoptions.GetOpt) getoptions.CommandFn

// NewTask - To be called within the TaskDefinitions function.
//
// The first argument is a task map that can be used to define task dependencies for DAGs.
// The second argument is the GetOpt instance that allows to nest commands into hierarchies.
//
// Example:
//
//	import (
//		"github.com/DavidGamba/dgtools/bake/lib/bake"
//		"github.com/DavidGamba/go-getoptions"
//		"github.com/DavidGamba/go-getoptions/dag"
//	)
//
//	func TaskDefinitions(ctx context.Context, opt *getoptions.GetOpt) error {
//		TM = dag.NewTaskMap()
//
//		m := bake.NewTask(TM, opt, "main", "main", Main)
//		bake.NewTask(TM, m, "hello", "This is a greeting", Hello)
//		bake.NewTask(TM, m, "world", "This is a planet", World)
//
//		return nil
//	}
func NewTask(tm *dag.TaskMap, opt *getoptions.GetOpt, name string, fn TaskFn) *getoptions.GetOpt {
	cmd := opt.NewCommand(name, "")
	fnr := fn(cmd)
	tm.Add(name, fnr)
	cmd.SetCommandFn(fnr)
	return cmd
}
