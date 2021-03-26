package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type ConfigCommand struct {
	Meta
}

func (f *ConfigCommand) Help() string {
	helpText := `
Usage: vaultcli config <subcommand> [options] [args]
  This command groups subcommands for interacting with Config system.
  Users can bootstrap vaultcli's Config system, create policies that restrict access,
  and generate tokens from those policies.
  Bootstrap Configs:
      $ vaultcli config bootstrap
  Please see the individual subcommand help for detailed usage information.
`
	return strings.TrimSpace(helpText)
}

func (f *ConfigCommand) Synopsis() string {
	return "Interact with Config"
}

func (f *ConfigCommand) Name() string { return "config" }

func (f *ConfigCommand) Run(args []string) int {
	return cli.RunResultHelp
}
