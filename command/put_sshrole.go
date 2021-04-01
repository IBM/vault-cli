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

type PutSSHRoleCommand struct {
	Meta Meta
}

func (c *PutSSHRoleCommand) Help() string {
	helpText := `
Usage: vault-cli put sshrole [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutSSHRoleCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutSSHRoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutSSHRoleCommand) Synopsis() string {
	return "put sshrole configures a role on a ssh endpoint"
}

func (c *PutSSHRoleCommand) Name() string { return "ssh role" }

func (c *PutSSHRoleCommand) Run(args []string) int {

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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/sshrole/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Policy (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/sshrole/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		yamlbytes, err := c.Meta.TemplateService.Exec("SSHRole", data, c.Meta.flagData)
		if err != nil {
			fmt.Printf("unable to apply template to sshrole: %s\n", err.Error())
			return 1
		}
		sshrole := vaultapi.SSHRole{}
		err = yaml.Unmarshal(yamlbytes, &sshrole)
		if err != nil {
			fmt.Printf("unable to marshal sshrole: %s\n", err.Error())
			return 1
		}

		c.Meta.SecretService.GetClient().SetNamespace(sshrole.Spec.VaultNamespace)

		name := sshrole.Spec.RoleName
		signerPath := sshrole.Spec.SignerPath

		sshiter := jsoniter.Config{TagKey: "vault"}.Froze()

		// unmarshal the Role Options
		data, err = sshiter.Marshal(sshrole.Spec.Parameters)
		m := make(map[string]interface{})
		json.Unmarshal(data, &m)

		_, err = c.Meta.SecretService.Write(fmt.Sprintf("/%s/roles/%s", signerPath, name), m)
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
			return 1
		}
		fmt.Printf("SSH Role (%s) write OK\n", f)
	}

	return 0
}
