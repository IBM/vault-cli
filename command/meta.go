package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ibm/vault-cli/pkg/config"
	"github.com/ibm/vault-cli/pkg/configservice"
	"github.com/ibm/vault-cli/pkg/secretservice"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/posener/complete"
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet.
type FlagSetFlags uint

// Meta contains the meta-options and functionality that nearly every
// Nomad command inherits.
type Meta struct {
	Ui cli.Ui

	flagConfigPath string
	Config         *config.Config
	ConfigService  configservice.ConfigService

	SecretService secretservice.SecretService

	CurrentContext *config.Context

	// // These are set by the command line flags.
	// context is the context name to use for this command
	currentContextName string
	// Whether to not-colorize output
	noColor bool

	// namespace to send API requests
	namespace     string
	name          string
	outputFormat  string
	InventoryPath string
}

// FlagSet returns a FlagSet with the common flags that every
// command implements. The exact behavior of FlagSet can be configured
// using the flags as the second parameter, for example to disable
// server settings on the commands that don't talk to a server.
func (m *Meta) FlagSet(n string) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	f.StringVar(&m.flagConfigPath, "config", "", "")
	f.StringVar(&m.currentContextName, "c", "local", "")
	f.StringVar(&m.currentContextName, "context", "local", "")
	f.StringVar(&m.namespace, "n", "", "")
	f.StringVar(&m.namespace, "namespace", "", "")
	f.StringVar(&m.outputFormat, "o", "", "")
	f.StringVar(&m.outputFormat, "output", "", "")

	f.SetOutput(&uiErrorWriter{ui: m.Ui})

	return f
}

// AutocompleteFlags returns a set of flag completions for the given flag set.
func (m *Meta) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-c":         complete.PredictAnything,
		"-context":   complete.PredictAnything,
		"-config":    complete.PredictAnything,
		"-n":         complete.PredictAnything,
		"-namespace": complete.PredictAnything,
		"-no-color":  complete.PredictNothing,
	}
}

// generalOptionsUsage returns the help string for the global options.
func generalOptionsUsage() string {

	helpText := `
  -context=<contextname>
    The name of the context to use for this run of the command
    Alias: -c
  -config=<vault-cli-config path>
    The location if the cli config yaml file. Defaults to "~/.vault-cli/config.yaml"

  -namespace=<namespace>
    The target namespace for queries and actions bound to a namespace.
    Overrides the VAULT_CLI_NAMESPACE environment variable if set.
    If set to '*', job and alloc subcommands query all namespaces authorized
    to user.
    Defaults to the "default" namespace.

  -no-color
    Disables colored command output. Alternatively, VAULT_CLI_NO_COLOR may be
    set.

  -output=<json|yaml|text>
    Alias: -o
`
	return strings.TrimSpace(helpText)
}

// funcVar is a type of flag that accepts a function that is the string given
// by the user.
type funcVar func(s string) error

func (f funcVar) Set(s string) error { return f(s) }
func (f funcVar) String() string     { return "" }
func (f funcVar) IsBoolFlag() bool   { return false }

func (m *Meta) Load() error {
	configPath, err := m.getConfigPath()
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting config path: %s\n", err.Error()))
	}
	cfg, err := m.ConfigService.Read(configPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading config: %s\n", err.Error()))
	}
	m.Config = cfg

	ctx := cfg.GetContextByName(m.currentContextName)
	if ctx == nil {
		return errors.New(fmt.Sprintf("could not find named context"))
	}
	m.CurrentContext = ctx

	secretsvc, err := m.Config.GetServiceFromContext(ctx, m.flagConfigPath, m.namespace)
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting service from config: %s\n", err.Error()))
	}
	m.SecretService = secretsvc
	return nil
}

// getConfigPath will set path based on:
// if set by flag override other methods
// if env variable set override default
// default
func (m *Meta) getConfigPath() (string, error) {
	var testDir string
	if m.flagConfigPath != "" { // testDir is set by flag
		testDir = m.flagConfigPath
	} else if testDir = os.Getenv(envVaultCLIConfigDir); testDir != "" { // testDir is specified in env
	} else { // create the default testDir
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		testDir = home + "/" + configDefaultDir
	}
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		err = os.Mkdir(testDir, 0755)
		if err != nil {
			return "", err
		}
	}
	path := testDir + "/" + configDefaultFileName
	return path, nil
}
