package command

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ibm/vault-cli/pkg/inventory"
	vaultapi "github.com/ibm/vault-go/api/v1"
	jsoniter "github.com/json-iterator/go"
	"github.com/posener/complete"
	"gopkg.in/yaml.v2"
)

type PutPKIRoleCommand struct {
	Meta Meta
}

func (c *PutPKIRoleCommand) Help() string {
	helpText := `
Usage: vault-cli put pkirole [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutPKIRoleCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutPKIRoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutPKIRoleCommand) Synopsis() string {
	return "put pkirole configures a role on a pki endpoint"
}

func (c *PutPKIRoleCommand) Name() string { return "pki role" }

func (c *PutPKIRoleCommand) Run(args []string) int {

	// get the flags specific to this command

	flagSet := c.Meta.FlagSet(c.Name())
	flagSet.Usage = func() { c.Meta.Ui.Output(c.Help()) }
	if err := flagSet.Parse(args); err != nil {
		return 1
	}

	// process args
	args = flagSet.Args()
	filespec := args[0]

	// load config
	err := c.Meta.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Meta Load error: %s\n", err.Error())
		return 1
	}

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/pkirole/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Policy (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/pkirole/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		pkirole := vaultapi.PKIRole{}
		err = yaml.Unmarshal(data, &pkirole)
		if err != nil {
			fmt.Printf("unable to marshal pkirole: %s\n", err.Error())
			return 1
		}

		c.Meta.SecretService.GetClient().SetNamespace(pkirole.Spec.VaultNamespace)

		pkiiter := jsoniter.Config{TagKey: "vault"}.Froze()

		// unmarshal the Role Options
		data, err = pkiiter.Marshal(pkirole.Spec.Config)
		m := make(map[string]interface{})
		json.Unmarshal(data, &m)
		_, err = c.Meta.SecretService.Write(fmt.Sprintf("/%s/roles/%s", pkirole.Spec.IssuerPath, pkirole.Spec.RoleName), m)
		if err != nil {
			//fmt.Printf(err.Error())
			fmt.Printf("(%s) %s", f, err)
			return 1
		}
		fmt.Printf("PKI Role (%s) write OK\n", f)
	}

	return 0
}
