package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	pkgargs "github.com/ibm/vault-cli/pkg/args"
	"github.com/ibm/vault-cli/pkg/inventory"
	vaultapi "github.com/ibm/vault-go/api/v1"
	"github.com/posener/complete"
	"gopkg.in/yaml.v2"
)

type PutSecretCommand struct {
	Meta  Meta
	ioDir string
}

func (c *PutSecretCommand) Help() string {
	helpText := `
Usage: vault-cli put secretmeta [options]

General Options:
  ` + generalOptionsUsage() + `
`
	return strings.TrimSpace(helpText)
}

func (c *PutSecretCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(),
		complete.Flags{})
}

func (c *PutSecretCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PutSecretCommand) Synopsis() string {
	return "put secretmeta writes a secret to vault conforming to the secretmeta spec"
}

func (c *PutSecretCommand) Name() string { return "secret" }

func (c *PutSecretCommand) Run(args []string) int {

	// get the flags specific to this command

	flagSet := c.Meta.FlagSet(c.Name())
	flagSet.StringVar(&c.ioDir, "dir", "", "")
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

	files, err := inventory.GetFiles(c.Meta.CurrentContext.InventoryPath+"/secretmeta/", filespec)
	if err != nil {
		fmt.Printf("get files error: %s\n", err.Error())
		return 1
	}
	if len(files) != 1 {
		fmt.Printf("SecretMeta (%s) not found in inventory", filespec)
		return 1
	}

	for _, f := range files {
		filename := c.Meta.CurrentContext.InventoryPath + "/secretmeta/" + f
		data, err := inventory.ReadFile(filename + ".yaml")
		if err != nil {
			fmt.Println("error reading file: ", err.Error())
			return 1
		}
		yamlbytes, err := c.Meta.TemplateService.Exec("Secret", data, c.Meta.flagData)
		if err != nil {
			fmt.Printf("unable to apply template to secretmeta: %s\n", err.Error())
			return 1
		}
		secretmeta := vaultapi.SecretMeta{}
		err = yaml.Unmarshal(yamlbytes, &secretmeta)
		if err != nil {
			fmt.Printf("unable to marshal secretmeta: %s\n", err.Error())
			return 1
		}
		if secretmeta.Spec.Type != "kv-v2" {
			fmt.Printf("secret type must be kv-v2\n")
			return 1
		}
		path := secretmeta.Spec.KVPath.Path
		if c.ioDir != "" {
			for _, key := range secretmeta.Spec.KVPath.Keys {
				filename := c.ioDir + string(os.PathSeparator) + key.Name
				if _, err := os.Stat(filename); err == nil {
					args = append(args, key.Name+"=@"+filename)
				}
			}
		}
		// Pull our fake stdin if needed
		stdin := (io.Reader)(os.Stdin)
		argArray, err := pkgargs.ParseArgsData(stdin, args[1:])
		if err != nil {
			fmt.Printf("Failed to parse K=V argArray: %s\n", err)
			return 1
		}
		// all keys defined in secretmeta must be present
		for _, k := range secretmeta.Spec.KVPath.Keys {
			if argArray[k.Name] == nil {
				fmt.Printf("required key not defined (key: %s)\n", k.Name)
				return 1
			}
		}
		// look for unknown or misspelled key name
		for k := range argArray {
			if _, ok := GetKeyFromKVKeysByName(secretmeta.Spec.KVPath.Keys, k); !ok {
				fmt.Printf("unknown key provided (key: %s)\n", k)
				return 1
			}
		}
		mountPath, v2, err := c.Meta.SecretService.IsKVv2(path)
		if err != nil {
			fmt.Printf("error:%s\n", err.Error())
			return 1
		}

		if v2 {
			path = pkgargs.AddPrefixToVKVPath(path, mountPath, "data")
			//path = mountPath + "data" + path
			argArray = map[string]interface{}{
				"data":    argArray,
				"options": map[string]interface{}{},
			}

			// if c.flagCAS > -1 {
			// 	data["options"].(map[string]interface{})["cas"] = c.flagCAS
			// }
		}
		secret, err := c.Meta.SecretService.Write(path, argArray)
		if err != nil {
			fmt.Printf("Error writing data to %s: %s\n", path, err)
			return 1
		}
		if secret != nil {
			out, err := json.MarshalIndent(secret, "", "  ")
			if err != nil {
				fmt.Printf("error marshaling secret :%s\n", err.Error())
				return 1
			}
			fmt.Println(string(out))
		}

	}

	return 0
}

// GetKeyFromKVKeysByName searches kvKey array for a key with the name
func GetKeyFromKVKeysByName(keys []vaultapi.KVKey, e string) (*vaultapi.KVKey, bool) {
	for _, a := range keys {
		if a.Name == e {
			return &a, true
		}
	}
	return nil, false
}
