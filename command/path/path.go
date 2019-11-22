package path

import (
	"fmt"

	"github.com/RTradeLtd/ca-cli/command"
	"github.com/RTradeLtd/ca-cli/config"
	"github.com/urfave/cli"
)

func init() {
	cmd := cli.Command{
		Name:        "path",
		Usage:       "print the configured step path and exit",
		UsageText:   "step path",
		Description: "**step ca** command prints the configured step path and exit",
		Action: cli.ActionFunc(func(ctx *cli.Context) error {
			fmt.Println(config.StepPath())
			return nil
		}),
	}

	command.Register(cmd)
}
