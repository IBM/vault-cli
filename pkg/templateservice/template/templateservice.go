package template

import (
	"bytes"
	"encoding/json"
	"html/template"

	"github.com/ibm/vault-cli/pkg/templateservice"
)

type templateService struct {
}

func MakeTemplateService() templateservice.TemplateService {
	return &templateService{}
}

func (t *templateService) Exec(name string, tpl []byte, data string) ([]byte, error) {
	if data == "" {
		data = "{}"
	}
	var yamlbytes []byte
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}
	var err error
	yamlbytes, err = t.ParseAndExecute(name, tpl, m)
	if err != nil {
		return nil, err
	}
	return yamlbytes, nil
}

func (t *templateService) ParseAndExecute(name string, tpl []byte, m map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, m)
	if err != nil {
		return nil, err
	}
	b := buf.Bytes()
	return b, nil
}
