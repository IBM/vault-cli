package command

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ibm/vault-cli/pkg/inventory"
	v1 "github.com/ibm/vault-go/api/v1"
	vaultapi "github.com/ibm/vault-go/api/v1"
	jsoniter "github.com/json-iterator/go"
	"github.com/posener/complete"
	"gopkg.in/yaml.v2"
)

type PutVaultEndpointCommand struct {
	Meta Meta
}

func (c *PutVaultEndpointCommand) Help() string {
	helpText := `
Usage: vault-cli put endpoint [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutVaultEndpointCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(), complete.Flags{
		"-force": complete.PredictAnything},
	)
}

func (c *PutVaultEndpointCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutVaultEndpointCommand) Synopsis() string {
	return "put endpoint configures an endpoint"
}

func (c *PutVaultEndpointCommand) Name() string { return "pki role" }

func (c *PutVaultEndpointCommand) Run(args []string) int {

	// get the flags specific to this command
	var putVaultEndpointForce bool
	flagSet := c.Meta.FlagSet(c.Name())
	flagSet.BoolVar(&putVaultEndpointForce, "force", false, "")
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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/vaultendpoint/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) == 0 {
		fmt.Printf("Vault Endpoint(%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/vaultendpoint/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		yamlbytes, err := c.Meta.TemplateService.Exec("VaultEndpoint", data, c.Meta.flagData)
		if err != nil {
			fmt.Printf("unable to apply template to vaultendpoint: %s\n", err.Error())
			return 1
		}
		endpoint := vaultapi.VaultEndpoint{}
		err = yaml.Unmarshal(yamlbytes, &endpoint)
		if err != nil {
			fmt.Printf("unable to marshal endpoint: %s\n", err.Error())
			return 1
		}

		c.Meta.SecretService.GetClient().SetNamespace(endpoint.Spec.VaultNamespace)
		// Mount vaultendpoint options
		pkiiter := jsoniter.Config{TagKey: "vault"}.Froze()

		// unmarshal the mountOptions
		data, err = pkiiter.Marshal(endpoint.Spec.MountOptions)
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
			return 1
		}
		m := make(map[string]interface{})
		json.Unmarshal(data, &m)

		endpointPreviouslyMounted := true
		_, err = c.Meta.SecretService.Read(fmt.Sprintf("sys/mounts/%s/tune", endpoint.Spec.Path))
		if err != nil {
			endpointPreviouslyMounted = false
			_, err = c.Meta.SecretService.Write(fmt.Sprintf("/sys/mounts/%s", endpoint.Spec.Path), m)
			if err != nil {
				fmt.Printf("(%s) %s", f, err)
			}
		}
		//		if endpoint.Spec.MountOptions.Type != "ssh" {
		data, err = pkiiter.Marshal(endpoint.Spec.TuneOptions)
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
		}
		m = make(map[string]interface{})
		err = json.Unmarshal(data, &m)
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
		}
		_, err = c.Meta.SecretService.Write(fmt.Sprintf("sys/mounts/%s/tune", endpoint.Spec.Path), m)
		if err != nil {
			fmt.Printf("(%s) %s", f, err)
		}

		//		}
		if endpoint.Spec.MountOptions.Type == "ssh" {
			if !endpointPreviouslyMounted {
				err = c.ConfigureSSHGenerateSigning(f, endpoint.Spec.Path, &endpoint)
				if err != nil {
					fmt.Printf("(%s) %s", f, err)
					return 1
				}
				fmt.Printf("SSH Endpoint configured (%s) write OK\n", f)
			}
		}

		// START PKI
		if endpoint.Spec.MountOptions.Type == "pki" {
			if !endpointPreviouslyMounted || putVaultEndpointForce {
				if endpoint.Spec.PKIConfig.RootOptions.GenerateOptions != (*v1.VaultGenerateOptions)(nil) {
					// TODO handle external Root CA
					if !endpoint.Spec.PKIConfig.ExportPrivateKey {
						err = c.ConfigureRootCAInternal(f, endpoint.Spec.Path, &endpoint)
						if err != nil {
							fmt.Printf("(%s) %s", f, err)
							return 1
						}
						fmt.Printf("PKI Root configured (%s) write OK\n", f)
					}
				}
				if endpoint.Spec.PKIConfig.IntermediateOptions.GenerateOptions != (*v1.VaultGenerateOptions)(nil) {
					// TODO handle external
					err = c.ConfigureIntermediateCAInternal(f, endpoint.Spec.Path, &endpoint)
					if err != nil {
						fmt.Printf("(%s) %s", f, err)
						return 1
					}
					fmt.Printf("PKI Intermediate configured (%s) write OK\n", f)
				}
				if endpoint.Spec.PKIConfig.URLs != (*v1.VaultEndpointConfigURLs)(nil) {
					err = c.ConfigureURLs(f, endpoint.Spec.Path, &endpoint)
					if err != nil {
						fmt.Printf("(%s) %s", f, err)
						return 1
					}
				}
				fmt.Printf("PKI Endpoint configured (%s) write OK\n", f)
			} else {
				fmt.Printf("PKI Endpoint already configured (%s)  SKIPPING\n", f)
			}
		}
		// End PKI
		fmt.Printf("Endpoint mount/tune (%s) write OK\n", f)

	}

	return 0
}

// ConfigureSSHGenerateSigning configures the endpoint
func (c *PutVaultEndpointCommand) ConfigureSSHGenerateSigning(filename, path string, endpoint *v1.VaultEndpoint) error {
	m := make(map[string]interface{})
	m["generate_signing_key"] = true
	_, err := c.Meta.SecretService.Write(fmt.Sprintf("/%s/config/ca", path), m)
	if err != nil {
		return fmt.Errorf("namespace: %s, (%s) %s", endpoint.Spec.VaultNamespace, filename, err)
	}
	return nil
}

// ConfigureRootCAInternal configures the endpoint
func (c *PutVaultEndpointCommand) ConfigureRootCAInternal(filename, path string, endpoint *v1.VaultEndpoint) error {
	pkiiter := jsoniter.Config{TagKey: "vault"}.Froze()

	data, err := pkiiter.Marshal(endpoint.Spec.PKIConfig.RootOptions.GenerateOptions)
	m := make(map[string]interface{})
	json.Unmarshal(data, &m)

	_, err = c.Meta.SecretService.Write(fmt.Sprintf("/%s/root/generate/internal", path), m)
	if err != nil {
		return fmt.Errorf("(%s) %s", filename, err)
	}

	return nil
}

// ConfigureIntermediateCAInternal configures the endpoint
func (c *PutVaultEndpointCommand) ConfigureIntermediateCAInternal(filename, intermediatePath string, endpoint *v1.VaultEndpoint) error {
	pkiiter := jsoniter.Config{TagKey: "vault"}.Froze()

	data, err := pkiiter.Marshal(endpoint.Spec.PKIConfig.IntermediateOptions.GenerateOptions)
	m := make(map[string]interface{})
	json.Unmarshal(data, &m)

	secret, err := c.Meta.SecretService.Write(fmt.Sprintf("/%s/intermediate/generate/internal", intermediatePath), m)
	if err != nil {
		return fmt.Errorf("(%s) %s", filename, err)
	}
	m["csr"] = secret.Data["csr"].(string)
	rootPath := endpoint.Spec.PKIConfig.IntermediateOptions.RootCAPath
	c.Meta.SecretService.GetClient().SetNamespace(endpoint.Spec.PKIConfig.IntermediateOptions.RootCANamespace)
	secret, err = c.Meta.SecretService.Write(fmt.Sprintf("/%s/root/sign-intermediate", rootPath), m)
	if err != nil {
		return fmt.Errorf("(%s) %s", filename, err)
	}
	if secret.Data == nil {
		return fmt.Errorf("error: expected certificate")
	}
	chain, err := c.Meta.SecretService.Read(fmt.Sprintf("/%s/cert/ca_chain", rootPath))
	if err != nil {
		return fmt.Errorf("(%s) %s", filename, err)
	}
	if chain.Data == nil {
		return fmt.Errorf("error: expected certificate")
	}

	c.Meta.SecretService.GetClient().SetNamespace(endpoint.Spec.VaultNamespace)
	m = make(map[string]interface{})
	if chain.Data["certificate"].(string) != "" {
		m["certificate"] = secret.Data["certificate"].(string) + "\n" + chain.Data["certificate"].(string)
		//		fmt.Print(m["certificate"])
	} else {
		m["certificate"] = secret.Data["certificate"].(string)
	}
	secret, err = c.Meta.SecretService.Write(fmt.Sprintf("/%s/intermediate/set-signed", intermediatePath), m)
	if err != nil {
		return fmt.Errorf("(%s) %s", filename, err)
	}
	return nil
}

// ConfigureURLs configures the endpoint
// TODO Resolve issues with where this information comes from
func (c *PutVaultEndpointCommand) ConfigureURLs(filename, path string, endpoint *v1.VaultEndpoint) error {
	return nil
}

// ConfigureCRLs configures the endpoint
// TODO Resolve issues with where this information comes from
func (c *PutVaultEndpointCommand) ConfigureCRLs(filename, path string, endpoint *v1.VaultEndpoint) error {
	return nil
}
