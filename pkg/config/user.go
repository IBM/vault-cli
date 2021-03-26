package config

import "errors"

// GetUserByName returns a named cluster
func (c *Config) GetUserByName(name string) *User {
	for _, user := range c.Users {
		if user.Name == name {
			return user
		}
	}
	return nil
}

// SetCertUser will create a new user or update an existing user
func (c *Config) SetCertUser(name string,
	clientCert string,
	clientCertData string,
	clientKey string,
	clientKeyData string,
) (*User, error) {
	found := &User{}
	if name == "" {
		return nil, errors.New("SetUser: name cannot be empty")
	}
	if found = c.GetUserByName(name); found == nil {

		newUser := User{
			Name: name,
			UserSpec: UserSpec{
				ClientCert:     clientCert,
				ClientCertData: clientCertData,
				ClientKey:      clientKey,
				ClientKeyData:  clientKeyData,
			},
		}
		if c.Users == nil {
			c.Users = []*User{}
		}
		c.Users = append(c.Users, &newUser)
		return &newUser, nil
	}
	if clientCert != "" {
		found.ClientCert = clientCert
	}
	if clientCertData != "" {
		found.ClientCertData = clientCertData
	}
	if clientKey != "" {
		found.ClientKey = clientKey
	}
	if clientKeyData != "" {
		found.ClientKeyData = clientKeyData
	}
	return found, nil
}

// SetUserPassUser will create a new user or update an existing user
func (c *Config) SetUserPassUser(name string,
	username string,
	password string,
) (*User, error) {
	found := &User{}
	if name == "" {
		return nil, errors.New("SetUser: name cannot be empty")
	}
	if found = c.GetUserByName(name); found == nil {

		newUser := User{
			Name: name,
			UserSpec: UserSpec{
				Username: username,
				Password: password,
			},
		}
		if c.Users == nil {
			c.Users = []*User{}
		}
		c.Users = append(c.Users, &newUser)
		return &newUser, nil
	}
	if username != "" {
		found.Username = username
	}
	if password != "" {
		found.Password = password
	}
	return found, nil
}

// SetAppRoleUser will create a new user or update an existing user
func (c *Config) SetAppRoleUser(name string,
	roleID string,
	secretID string,
) (*User, error) {
	found := &User{}
	if name == "" {
		return nil, errors.New("SetUser: name cannot be empty")
	}
	if found = c.GetUserByName(name); found == nil {

		newUser := User{
			Name: name,
			UserSpec: UserSpec{
				RoleID:   roleID,
				SecretID: secretID,
			},
		}
		if c.Users == nil {
			c.Users = []*User{}
		}
		c.Users = append(c.Users, &newUser)
		return &newUser, nil
	}
	if roleID != "" {
		found.RoleID = roleID
	}
	if secretID != "" {
		found.SecretID = secretID
	}
	return found, nil
}

// DeleteUser will delete the named cluster
func (c *Config) DeleteUser(name string) error {
	if found := c.GetUserByName(name); found != nil {
		// should check if this user is in the current user
		ctx := c.GetContextByName(c.CurrentContext)
		if ctx.User == name {
			return errors.New("cannot delete user in current context")
		}
		users := []*User{}
		for _, user := range c.Users {
			if user.Name != found.Name {
				users = append(users, user)
			}
		}
		c.Users = users
		return nil
	}
	return errors.New("user not found")
}
