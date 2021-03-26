package config

import (
	"errors"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/ibm/vault-cli/pkg/inventory"
	"github.com/ibm/vault-cli/pkg/secretservice"
	"github.com/ibm/vault-cli/pkg/secretservice/vault"
)

// GetDefaultClient attempt to login ot vault
func (cfg *Config) GetDefaultClient(namespace, url, token, cert, key, cacert string) (*api.Client, error) {
	//token = VaultLogin(namespace, url, cert, key, cacert)
	os.Setenv("VAULT_ADDR", url)
	os.Setenv("VAULT_CLIENT_CERT", inventory.ExpandHomePath(cert))
	os.Setenv("VAULT_CLIENT_KEY", inventory.ExpandHomePath(key))
	os.Setenv("VAULT_CACERT", inventory.ExpandHomePath(cacert))
	os.Setenv("VAULT_TOKEN", token)
	apiCfg := api.DefaultConfig()
	apiClient, err := api.NewClient(apiCfg)
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		apiClient.SetNamespace(namespace)
	}
	return apiClient, nil
}

// GetClientFromContext gets user/cluster/namespace info from context
func (cfg *Config) GetClientFromContext(secretsvc secretservice.SecretService, configfile, contextName, namespace string) (*api.Client, error) {
	ctx := cfg.GetContextByName(contextName)

	if ctx == nil {
		return nil, errors.New("could not find named context")
	}

	cluster := cfg.GetClusterByName(ctx.Cluster)
	user := cfg.GetUserByName(ctx.User)
	// if cluster == nil || cluster.Server == "" || cluster.CertAuth == "" {
	// 	return nil, errors.New("cluster must have server address and CA cert")
	// }
	// if user == nil || user.ClientCert == "" || user.ClientKey == "" {
	// 	return nil, errors.New("user must have cert and key")
	// }
	session, err := cfg.GetSession(secretsvc, configfile, contextName, false)
	if err != nil {
		return nil, err
	}
	ns := ctx.Namespace
	if namespace != "" {
		ns = namespace
	}
	client, err := cfg.GetDefaultClient(ns, cluster.Server, session.Token, user.ClientCert, user.ClientKey, cluster.CertAuth)
	// override environment variable
	if ns != "" {
		if ns == "root" {
			ns = ""
		}
		client.SetNamespace(ns)
	}
	return client, nil
}

// GetServiceFromContext gets user/cluster/namespace info from context
func (cfg *Config) GetServiceFromContext(ctx *Context, configfile, namespace string) (secretservice.SecretService, error) {
	cluster := cfg.GetClusterByName(ctx.Cluster)
	user := cfg.GetUserByName(ctx.User)
	// if cluster == nil || cluster.Server == "" || cluster.CertAuth == "" {
	// 	return nil, errors.New("cluster must have server address and CA cert")
	// }
	// if user == nil || user.ClientCert == "" || user.ClientKey == "" {
	// 	return nil, errors.New("user must have cert and key")
	// }
	secretsvc := vault.NewVaultService()
	session, err := cfg.GetSession(secretsvc, configfile, ctx.Name, false)
	if err != nil {
		return nil, err
	}
	ns := ctx.Namespace
	if namespace != "" {
		ns = namespace
	}
	// This is going to be returned as a vaultstore.Store interface
	defaultClient, err := cfg.GetDefaultClient(ns, cluster.Server, session.Token, user.ClientCert, user.ClientKey, cluster.CertAuth)
	// override environment variable
	if ns != "" {
		if ns == "root" {
			ns = ""
		}
		defaultClient.SetNamespace(ns)
	}
	secretsvc.SetClient(defaultClient)
	return secretsvc, nil
}
