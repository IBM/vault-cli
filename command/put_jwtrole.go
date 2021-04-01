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

type PutJWTRoleCommand struct {
	Meta Meta
}

func (c *PutJWTRoleCommand) Help() string {
	helpText := `
Usage: vault-cli put jwtrole [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutJWTRoleCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutJWTRoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutJWTRoleCommand) Synopsis() string {
	return "put jwtrole configures a role on a jwt endpoint"
}

func (c *PutJWTRoleCommand) Name() string { return "jwt role" }

func (c *PutJWTRoleCommand) Run(args []string) int {

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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/jwtrole/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Policy (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/jwtrole/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		yamlbytes, err := c.Meta.TemplateService.Exec("JWTRole", data, c.Meta.flagData)
		if err != nil {
			fmt.Printf("unable to apply template to jwtrole: %s\n", err.Error())
			return 1
		}
		jwtrole := vaultapi.JWTRole{}
		err = yaml.Unmarshal(yamlbytes, &jwtrole)
		if err != nil {
			fmt.Printf("unable to marshal jwtrole: %s\n", err.Error())
			return 1
		}

		c.Meta.SecretService.GetClient().SetNamespace(jwtrole.Spec.VaultNamespace)

		jwtiter := jsoniter.Config{TagKey: "vault"}.Froze()

		// unmarshal the Role Options
		data, err = jwtiter.Marshal(jwtrole.Spec.Parameters)
		m := make(map[string]interface{})
		json.Unmarshal(data, &m)
		_, err = c.Meta.SecretService.Write(fmt.Sprintf("/auth/%s/role/%s", jwtrole.Spec.AuthPath, jwtrole.Spec.RoleName), m)
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
			return 1
		}
		fmt.Printf("JWT Role (%s) write OK\n", f)
	}

	return 0
}
