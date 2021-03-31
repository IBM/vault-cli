package templateservice

//go:generate counterfeiter -o fakes/templateservice.go --fake-name FakeTemplateService . TemplateService
type TemplateService interface {
	Exec(name string, tpl []byte, data string) ([]byte, error)
	ParseAndExecute(name string, tpl []byte, m map[string]interface{}) ([]byte, error)
}
