package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/ibm/vault-cli/pkg/secretservice"
)

// SetSession Builds a session object
// expireSkewFactor in seconds
func (c *Config) SetSession(token string, duration *int64, renewable *bool, expireSkewFactor int64) *Session {
	if token != "" || duration != nil || renewable != nil {
		session := Session{
			Token:         token,
			LeaseDuration: duration,
			Renewable:     renewable,
		}
		// reduce the expire time by 30 minutes
		zero := int64(0)
		if duration != nil && *duration > zero {
			expire := (time.Now().UTC().Unix() + int64(*duration)) - expireSkewFactor
			session.Expires = &expire
		}
		return &session
	}
	return nil
}

// vaultSessionExpireSkewFactor the amount of time to subtract from Expire to account for clock skew
const vaultSessionExpireSkewFactor = int64(30 * 60)

// GetSession will return an existing session or create a new one and save the config
func (cfg *Config) GetSession(secretsvc secretservice.SecretService, configfile, contextName string, forceNewSession bool) (*Session, error) {
	for _, c := range cfg.Contexts {
		if c.Name == contextName {
			now := time.Now().UTC().Unix()
			if c.Session.Expires == nil {
				zero := int64(0)
				c.Session.Expires = &zero
			}

			if forceNewSession || c.Session.Token == "" || now > *c.Session.Expires {
				cluster := cfg.GetClusterByName(c.Cluster)
				user := cfg.GetUserByName(c.User)
				if cluster == nil || cluster.Server == "" {
					return nil, errors.New("cluster must have server address")
				}
				if user == nil {
					return nil, errors.New("user must have cert and key")
				}
				var err error
				var response *api.Secret
				ns := c.Namespace
				if user.IgnoreNamespaceOnAuth == true {
					ns = ""
				}
				if user.ClientCert != "" {
					response, err = secretsvc.CertLogin(ns, cluster.Server, "cert", user.ClientCert, user.ClientKey, cluster.CertAuth, cluster.InsecureSkipTLSVerify)
				} else if user.Username != "" {
					response, err = secretsvc.UserPassLogin(ns, cluster.Server, "userpass", user.Username, user.Password, cluster.CertAuth, cluster.InsecureSkipTLSVerify)
				} else if user.RoleID != "" {
					response, err = secretsvc.AppRoleLogin(ns, cluster.Server, "approle", user.RoleID, user.SecretID, cluster.CertAuth, cluster.InsecureSkipTLSVerify)

				} else {
					return nil, fmt.Errorf("GetSession login requires credentials")
				}
				if err != nil {
					return nil, err
				}

				if response.Auth != nil {
					duration := int64(response.Auth.LeaseDuration)
					session := cfg.SetSession(response.Auth.ClientToken, &duration, &response.Auth.Renewable, vaultSessionExpireSkewFactor)
					c.Session = *session
					cfg.SaveConfig(configfile)
				}
			}
			return &c.Session, nil
		}
	}

	return nil, errors.New("context not found")
}

// StopSession will return an existing session or create a new one and save the config
func (cfg *Config) StopSession(configfile, contextName string) error {

	for _, c := range cfg.Contexts {
		if c.Name == contextName {
			c.Session = Session{}
		}
	}
	return nil
}
