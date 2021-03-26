package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type PutCommand struct {
	Meta
}

func (f *PutCommand) Help() string {
	helpText := `
Usage: nomad Put <subcommand> [options] [args]
  This command groups subcommands for interacting with Put policies and tokens.
  Users can bootstrap Nomad's Put system, create policies that restrict access,
  and generate tokens from those policies.
  Bootstrap Puts:
      $ nomad Put bootstrap
  Please see the individual subcommand help for detailed usage information.
`
	return strings.TrimSpace(helpText)
}

func (f *PutCommand) Synopsis() string {
	return "Interact with Put policies and tokens"
}

func (f *PutCommand) Name() string { return "Put" }

func (f *PutCommand) Run(args []string) int {
	return cli.RunResultHelp
}
