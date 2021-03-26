package configservice

import (
	"github.com/ibm/vault-cli/pkg/config"
)

// ConfigService Interface for controller operations needed by task workers
//go:generate counterfeiter -o fakes/configservice.go --fake-name FakeConfigService . ConfigService
type ConfigService interface {
	Read(path string) (*config.Config, error)
	Write(path string, cfg *config.Config) error
}
