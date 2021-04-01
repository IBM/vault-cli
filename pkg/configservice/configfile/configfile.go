package configfile

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/ibm/vault-cli/pkg/config"
	"github.com/ibm/vault-cli/pkg/configservice"
)

type configfile struct {
}

// NewConfigFileService should return a pointer to a configfile client
func NewConfigFileService() configservice.ConfigService {
	return &configfile{}
}

func (cf *configfile) Read(path string) (*config.Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil && strings.Contains(err.Error(), "no such file") {
		bytes = getDefaultConfig()
	} else if err != nil {
		return nil, err
	}
	c := config.Config{}
	err = yaml.Unmarshal([]byte(bytes), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (cf *configfile) Write(path string, cfg *config.Config) error {
	return nil
}

func getDefaultConfig() []byte {
	bytes := []byte(`apiVersion: v1
kind: Config
contexts:
- name: ns-test
  context:
    cluster: local
    inventoryPath: "hack/sample/ns-test"
    namespace: nextgen
    session:
      token: root
      lease-duration: 7200
      expires: 2582395696
      renewable: true
    user: localuser
- name: tpl-test
  context:
    cluster: local
    inventoryPath: "hack/sample/tpl-test"
    namespace: root
    session:
      token: root
      lease-duration: 7200
      expires: 2582395696
      renewable: true
    user: localuser
clusters:
- name: local
  cluster:
    certificate-authority: ""
    insecure-skip-tls-verify: true
    server: http://127.0.0.1:8200
current-context: local
users:
- name: localuser
  user:
    client-certificate: ""
    client-key: ""
    password: ""
    username: ""
    roleID: ""
    secretID: ""
    ignore-namespace-on-auth: false
`)
	return bytes

}
