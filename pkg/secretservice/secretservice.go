package secretservice

import (
	"github.com/hashicorp/vault/api"
)

// SecretService Interface for controller operations needed by task workers
//go:generate counterfeiter -o fakes/secretservice.go --fake-name FakeSecretService . SecretService
type SecretService interface {
	List(path string) (*api.Secret, error)
	Read(path string) (*api.Secret, error)
	ReadWithData(path string, data map[string][]string) (*api.Secret, error)
	Write(path string, data map[string]interface{}) (*api.Secret, error)
	Delete(path string) (*api.Secret, error)
	IsKVv2(path string) (string, bool, error)
	GetClient() *api.Client
	SetClient(c *api.Client)
	AppRoleLogin(namespace, authurl, endpoint, roleID, secretID, cacert string, insecureSkipVerify bool) (*api.Secret, error)
	CertLogin(namespace, url, endpoint, cert, key, cacert string, insecureSkipVerify bool) (*api.Secret, error)
	UserPassLogin(namespace, authurl, endpoint, username, password, cacert string, insecureSkipVerify bool) (*api.Secret, error)
}
