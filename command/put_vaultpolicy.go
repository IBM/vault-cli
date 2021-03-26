package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/ibm/vault-cli/pkg/inventory"
	vaultapi "github.com/ibm/vault-go/api/v1"
	"github.com/posener/complete"
	"github.com/rodaine/hclencoder"
	"gopkg.in/yaml.v2"
)

type PutVaultPolicyCommand struct {
	Meta Meta
}

func (c *PutVaultPolicyCommand) Help() string {
	helpText := `
Usage: vault-cli put vaultpolicy [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutVaultPolicyCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutVaultPolicyCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutVaultPolicyCommand) Synopsis() string {
	return "Bootstrap the ACL system for initial token"
}

func (c *PutVaultPolicyCommand) Name() string { return "acl bootstrap" }

func (c *PutVaultPolicyCommand) Run(args []string) int {

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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/vaultpolicy/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Policy (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/vaultpolicy/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		vaultPolicy := vaultapi.VaultPolicy{}
		err = yaml.Unmarshal(data, &vaultPolicy)
		if err != nil {
			fmt.Printf("unable to marshal vaultpolicy: %s\n", err.Error())
			return 1
		}

		c.Meta.SecretService.GetClient().SetNamespace(vaultPolicy.Spec.VaultNamespace)

		hcl, err := hclencoder.Encode(vaultPolicy.Spec.Policies)
		if err != nil {
			fmt.Printf("unable to encode: %s", err)
			return 1
		}
		m := make(map[string]interface{})
		strHCL := string(hcl)
		m["policy"] = strHCL

		_, err = c.Meta.SecretService.Write(fmt.Sprintf("sys/policy/%s", vaultPolicy.Spec.PolicyName), m)
		if err != nil {
			//cmd.Println(err)
			fmt.Printf("%v", err)
			return 1
		}
		fmt.Printf("Policy: %s.yaml, Name: %s, write, OK\n", filename, vaultPolicy.Spec.PolicyName)
	}

	return 0
}
