package config

import "errors"

// GetClusterByName returns a named cluster
func (c *Config) GetClusterByName(name string) *Cluster {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return cluster
		}
	}
	return nil
}

// SetCluster returns a new Cluster
func (c *Config) SetCluster(name string,
	certAuth string,
	certAuthData string,
	insecureSkipTLSVerify bool,
	server string,
) (*Cluster, error) {
	found := &Cluster{}
	if name == "" {
		return nil, errors.New("SetCluster: name cannot be empty")
	}
	if found = c.GetClusterByName(name); found == nil {
		nc := Cluster{
			Name: name,
			ClusterSpec: ClusterSpec{
				CertAuth:              certAuth,
				CertAuthData:          certAuthData,
				InsecureSkipTLSVerify: insecureSkipTLSVerify,
				Server:                server,
			},
		}
		c.Clusters = append(c.Clusters, &nc)
		return &nc, nil
	}
	if certAuth != "" {
		found.CertAuth = certAuth
	}
	if certAuthData != "" {
		found.CertAuthData = certAuthData
	}
	if insecureSkipTLSVerify != found.InsecureSkipTLSVerify {
		found.InsecureSkipTLSVerify = insecureSkipTLSVerify
	}
	if server != "" {
		found.Server = server
	}
	return found, nil
}

// DeleteCluster will delete the named cluster
func (c *Config) DeleteCluster(name string) error {
	if found := c.GetClusterByName(name); found != nil {
		ctx := c.GetContextByName(c.CurrentContext)
		if ctx.Cluster == name {
			return errors.New("cannot delete cluster in current context")
		}
		clusters := []*Cluster{}
		for _, cluster := range c.Clusters {
			if cluster.Name != name {
				clusters = append(clusters, cluster)
			}
		}
		c.Clusters = clusters
	}
	return nil
}
