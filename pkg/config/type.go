package config

type ClusterSpec struct {
	CertAuth              string `mapstructure:"certificate-authority" json:"certificate-authority" yaml:"certificate-authority"`
	CertAuthData          string `mapstructure:"certificate-authority-data,omitempty" json:"certificate-authority-data,omitempty" yaml:"certificate-authority-data,omitempty"`
	InsecureSkipTLSVerify bool   `mapstructure:"insecure-skip-tls-verify,omitempty" json:"insecure-skip-tls-verify,omitempty" yaml:"insecure-skip-tls-verify,omitempty"`
	Server                string `mapstructure:"server" json:"server" yaml:"server"`
}

// Cluster is a cluster definition with a name
type Cluster struct {
	Name        string `mapstructure:"name" json:"name" yaml:"name"`
	ClusterSpec `mapstructure:"cluster" json:"cluster" yaml:"cluster"`
}

// Session is a Token and its metadata
type Session struct {
	Token         string `mapstructure:"token,omitempty" json:"token,omitempty" yaml:"token,omitempty"`
	LeaseDuration *int64 `mapstructure:"lease-duration,omitempty" json:"lease-duration,omitempty" yaml:"lease-duration,omitempty"`
	Expires       *int64 `mapstructure:"expires,omitempty" json:"expires,omitempty" yaml:"expires,omitempty"`
	Renewable     *bool  `mapstructure:"renewable,omitempty" json:"renewable,omitempty" yaml:"renewable,omitempty"`
}

// ContextSpec consist of a user paired with a cluster, namespace
type ContextSpec struct {
	Cluster       string `mapstructure:"cluster" json:"cluster" yaml:"cluster"`
	InventoryPath string `mapstructure:"inventoryPath" json:"inventoryPath" yaml:"inventoryPath"`
	Namespace     string `mapstructure:"namespace" json:"namespace" yaml:"namespace"`
	Session       `mapstructure:"session,omitempty" json:"session,omitempty" yaml:"session,omitempty"`
	User          string `mapstructure:"user" json:"user" yaml:"user"`
}

// Context a context with a name
// +k8s:deepcopy-gen=true
type Context struct {
	Name        string `mapstructure:"name" json:"name" yaml:"name"`
	ContextSpec `mapstructure:"context" json:"context" yaml:"context"`
}

// UserSpec a user with credentials
type UserSpec struct {
	ClientCert            string `mapstructure:"client-certificate" json:"client-certificate" yaml:"client-certificate"`
	ClientCertData        string `mapstructure:"client-certificate-data,omitempty" json:"client-certificate-data,omitempty" yaml:"client-certificate-data,omitempty"`
	ClientKey             string `mapstructure:"client-key" json:"client-key" yaml:"client-key"`
	ClientKeyData         string `mapstructure:"client-key-data,omitempty" json:"client-key-data,omitempty" yaml:"client-key-data,omitempty"`
	Password              string `mapstructure:"password" json:"password" yaml:"password"`
	Username              string `mapstructure:"username" json:"username" yaml:"username"`
	RoleID                string `mapstructure:"roleID" json:"roleID" yaml:"roleID"`
	SecretID              string `mapstructure:"secretID" json:"secretID" yaml:"secretID"`
	IgnoreNamespaceOnAuth bool   `mapstructure:"ignore-namespace-on-auth" json:"ignore-namespace-on-auth" yaml:"ignore-namespace-on-auth"`
}

// User is a user with a name
type User struct {
	Name     string `mapstructure:"name" json:"name" yaml:"name"`
	UserSpec `mapstructure:"user" json:"user" yaml:"user"`
}

// Config defines config options
type Config struct {
	APIVersion     string     `mapstructure:"apiVersion" json:"apiVersion" yaml:"apiVersion"`
	Kind           string     `mapstructure:"kind" json:"kind" yaml:"kind"`
	Contexts       []*Context `mapstructure:"contexts" json:"contexts" yaml:"contexts"`
	Clusters       []*Cluster `mapstructure:"clusters" json:"clusters" yaml:"clusters"`
	CurrentContext string     `mapstructure:"current-context" json:"current-context" yaml:"current-context"`
	Users          []*User    `mapstructure:"users" json:"users" yaml:"users"`
}
