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

type PutVaultAuthCommand struct {
	Meta Meta
}

func (c *PutVaultAuthCommand) Help() string {
	helpText := `
Usage: vault-cli put auth [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutVaultAuthCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutVaultAuthCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutVaultAuthCommand) Synopsis() string {
	return "Create auth endpoint(s)"
}

func (c *PutVaultAuthCommand) Name() string { return "put auth" }

func (c *PutVaultAuthCommand) Run(args []string) int {

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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/vaultauth/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("VaultAuth (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/vaultauth/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		yamlbytes, err := c.Meta.TemplateService.Exec("VaultAuth", data, c.Meta.flagData)
		if err != nil {
			fmt.Printf("unable to apply template to vaultauth: %s\n", err.Error())
			return 1
		}
		vaultAuth := vaultapi.VaultAuth{}
		err = yaml.Unmarshal(yamlbytes, &vaultAuth)
		if err != nil {
			fmt.Printf("unable to marshal vaultauth: %s\n", err.Error())
			return 1
		}

		pkiiter := jsoniter.Config{TagKey: "vault"}.Froze()

		// unmarshal the mountOptions
		data, err = pkiiter.Marshal(vaultAuth.Spec.Data)
		m := make(map[string]interface{})
		json.Unmarshal(data, &m)

		if vaultAuth.Spec.VaultNamespace != "" {
			c.Meta.SecretService.GetClient().SetNamespace(vaultAuth.Spec.DeepCopy().VaultNamespace)
		}

		_, err = c.Meta.SecretService.Write(fmt.Sprintf("sys/auth/%s", vaultAuth.Spec.Path), m)
		if err != nil && strings.Contains(err.Error(), "path is already in use") {
			_, err = c.Meta.SecretService.Write(fmt.Sprintf("sys/auth/%s/tune", vaultAuth.Spec.Path), m)
		}
		if vaultAuth.Spec.Data.Type == "jwt" {
			data, err = pkiiter.Marshal(vaultAuth.Spec.JWTConfig)
			m := make(map[string]interface{})
			json.Unmarshal(data, &m)
			_, err = c.Meta.SecretService.Write(fmt.Sprintf("auth/%s/config", vaultAuth.Spec.Path), m)
			if err != nil {
				_, err = c.Meta.SecretService.Write(fmt.Sprintf("auth/%s/config", vaultAuth.Spec.Path), m)
			}

		}
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
			return 1
		}
		fmt.Printf("VaultAuth: %s.yaml, Name: %s write OK\n", f, vaultAuth.Spec.Path)

	}

	return 0
}
