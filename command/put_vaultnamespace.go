package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/ibm/vault-cli/pkg/inventory"
	vaultapi "github.com/ibm/vault-go/api/v1"
	"github.com/posener/complete"
	"gopkg.in/yaml.v2"
)

type PutVaultNamespaceCommand struct {
	Meta Meta
}

func (c *PutVaultNamespaceCommand) Help() string {
	helpText := `
Usage: vault-cli put vaultnamespace [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutVaultNamespaceCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutVaultNamespaceCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutVaultNamespaceCommand) Synopsis() string {
	return "Bootstrap the ACL system for initial token"
}

func (c *PutVaultNamespaceCommand) Name() string { return "acl bootstrap" }

func (c *PutVaultNamespaceCommand) Run(args []string) int {

	// get the flags specific to this command

	flagSet := c.Meta.FlagSet(c.Name())
	flagSet.Usage = func() { c.Meta.Ui.Output(c.Help()) }
	if err := flagSet.Parse(args); err != nil {
		return 1
	}

	// process args
	args = flagSet.Args()
	vaultnamespacefilespec := args[0]

	// load config
	err := c.Meta.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Meta Load error: %s\n", err.Error())
		return 1
	}

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/vaultnamespace/", vaultnamespacefilespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Namespace (%s) not found in inventory", vaultnamespacefilespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/vaultnamespace/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		vaultNamespace := vaultapi.VaultNamespace{}
		err = yaml.Unmarshal(data, &vaultNamespace)
		if err != nil {
			fmt.Printf("unable to marshal vaultnamespace: %s\n", err.Error())
			return 1
		}

		if vaultNamespace.Spec.NamespaceBase != "" {
			c.Meta.SecretService.GetClient().SetNamespace(vaultNamespace.Spec.NamespaceBase)
		}

		secret, err := c.Meta.SecretService.Read(fmt.Sprintf("/sys/namespaces/%s", vaultNamespace.Spec.NamespaceName))
		if err == nil && secret != nil {
			fmt.Printf("Vault Namespace: (%s.yaml) %s exists\n", f, vaultNamespace.Spec.NamespaceName)
			continue
		}
		m := make(map[string]interface{})
		_, err = c.Meta.SecretService.Write(fmt.Sprintf("/sys/namespaces/%s", vaultNamespace.Spec.NamespaceName), m)
		if err != nil {
			fmt.Printf("Vault Namespace: (%s.yaml) %s %s\n", f, vaultNamespace.Spec.NamespaceName, err)
			return 1
		}
		fmt.Printf("Vault Namespace: (%s.yaml) %s write, OK\n", f, vaultNamespace.Spec.NamespaceName)
	}

	return 0
}
