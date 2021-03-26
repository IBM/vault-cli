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

type PutVaultRoleCommand struct {
	Meta                         Meta
	FlagPolicies                 string
	FlagBoundNamespaces          string
	FlagBoundServiceAccountNames string
}

func (c *PutVaultRoleCommand) Help() string {
	helpText := `
Usage: vault-cli put vaultrole [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutVaultRoleCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutVaultRoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutVaultRoleCommand) Synopsis() string {
	return "Bootstrap the ACL system for initial token"
}

func (c *PutVaultRoleCommand) Name() string { return "acl bootstrap" }

func (c *PutVaultRoleCommand) Run(args []string) int {

	// get the flags specific to this command

	flagSet := c.Meta.FlagSet(c.Name())
	flagSet.StringVar(&c.FlagPolicies, "policies", "", "")
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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/vaultrole/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Role (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/vaultrole/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		vaultRole := vaultapi.VaultRole{}
		err = yaml.Unmarshal(data, &vaultRole)
		if err != nil {
			fmt.Printf("unable to marshal vaultrole: %s\n", err.Error())
			return 1
		}

		authMethod := vaultRole.Spec.AuthMethod
		roleName := vaultRole.Spec.RoleName
		c.Meta.SecretService.GetClient().SetNamespace(vaultRole.Spec.VaultNamespace)

		if c.FlagPolicies != "" {
			pols := strings.Split(c.FlagPolicies, ",")
			for _, v := range pols {
				vaultRole.Spec.Data.Policies = append(vaultRole.Spec.Data.Policies, v)
				vaultRole.Spec.Data.TokenPolicies = append(vaultRole.Spec.Data.TokenPolicies, v)
			}
		}
		if c.FlagBoundNamespaces != "" {
			pols := strings.Split(c.FlagBoundNamespaces, ",")
			for _, v := range pols {
				vaultRole.Spec.Data.BoundServiceAccountNamespaces = append(vaultRole.Spec.Data.BoundServiceAccountNamespaces, v)
			}
		}
		if c.FlagBoundServiceAccountNames != "" {
			bsans := strings.Split(c.FlagBoundServiceAccountNames, ",")
			for _, v := range bsans {
				vaultRole.Spec.Data.BoundServiceAccountNames = append(vaultRole.Spec.Data.BoundServiceAccountNames, v)
			}
		}

		// unmarshal the data
		pkiiter := jsoniter.Config{TagKey: "vault"}.Froze()
		data, err = pkiiter.Marshal(vaultRole.Spec.Data)
		m := make(map[string]interface{})
		json.Unmarshal(data, &m)

		_, err = c.Meta.SecretService.Write(fmt.Sprintf("auth/%s/role/%s", authMethod, roleName), m)
		if err != nil {
			fmt.Printf("Role (%s) %s", filename, err)
			return 1
		}
		fmt.Printf(string("Role: %s.yaml, Method: %s, Name: %s  write OK\n"), filename, authMethod, roleName)

	}

	return 0
}
