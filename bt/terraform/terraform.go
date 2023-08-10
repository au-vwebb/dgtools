package terraform

import (
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func NewCommand(parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("terraform", "terraform related tasks")
	initCMD(opt)
	planCMD(opt)
	applyCMD(opt)

	return opt
}
