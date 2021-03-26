package command

import (
	"os"

	colorable "github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
)

const (
	// EnvNomadCLINoColor is an env var that toggles colored UI output.
	EnvVaultCLINoColor = `VAULT_CLI_NO_COLOR`
)

// NamedCommand is a interface to denote a commmand's name.
type NamedCommand interface {
	Name() string
}

// Commands returns the mapping of CLI commands for Nomad. The meta
// parameter lets you set meta options for all commands.
func Commands(metaPtr *Meta, agentUI cli.Ui) map[string]cli.CommandFactory {
	if metaPtr == nil {
		metaPtr = new(Meta)
	}

	meta := *metaPtr
	if meta.Ui == nil {
		meta.Ui = &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      colorable.NewColorableStdout(),
			ErrorWriter: colorable.NewColorableStderr(),
		}
	}

	all := map[string]cli.CommandFactory{
		"config": func() (cli.Command, error) {
			return &ConfigCommand{
				Meta: meta,
			}, nil
		},
		"put": func() (cli.Command, error) {
			return &PutCommand{
				Meta: meta,
			}, nil
		},
		"put jwtrole": func() (cli.Command, error) {
			return &PutJWTRoleCommand{
				Meta: meta,
			}, nil
		},
		"put pkirole": func() (cli.Command, error) {
			return &PutPKIRoleCommand{
				Meta: meta,
			}, nil
		},
		"put sshrole": func() (cli.Command, error) {
			return &PutSSHRoleCommand{
				Meta: meta,
			}, nil
		},
		"put vaultauth": func() (cli.Command, error) {
			return &PutVaultAuthCommand{
				Meta: meta,
			}, nil
		},
		"put vaultendpoint": func() (cli.Command, error) {
			return &PutVaultEndpointCommand{
				Meta: meta,
			}, nil
		},
		"put vaultnamespace": func() (cli.Command, error) {
			return &PutVaultNamespaceCommand{
				Meta: meta,
			}, nil
		},
		"put vaultpolicy": func() (cli.Command, error) {
			return &PutVaultPolicyCommand{
				Meta: meta,
			}, nil
		},
		"put vaultrole": func() (cli.Command, error) {
			return &PutVaultRoleCommand{
				Meta: meta,
			}, nil
		},
	}

	for k, v := range EntCommands(metaPtr, agentUI) {
		all[k] = v
	}

	return all
}
