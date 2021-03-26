package config

import "errors"

// GetContextByName returns a named cluster
func (c *Config) GetContextByName(name string) *Context {
	for _, context := range c.Contexts {
		if context.Name == name {
			return context
		}
	}
	return nil
}

// SetContext will create a new context or update an existing context
func (c *Config) SetContext(name string,
	cluster string,
	namespace string,
	session *Session,
	user string) (*Context, error) {
	found := &Context{}
	if name == "" {
		return nil, errors.New("SetContext: name cannot be empty")
	}
	if found = c.GetContextByName(name); found == nil {

		newCtx := Context{
			Name: name,
			ContextSpec: ContextSpec{
				Cluster:   cluster,
				Namespace: namespace,
				Session:   *session,
				User:      user,
			},
		}
		if c.Contexts == nil {
			c.Contexts = []*Context{}
		}
		c.Contexts = append(c.Contexts, &newCtx)
		return &newCtx, nil
	}
	if cluster != "" {
		found.Cluster = cluster
	}
	if namespace != "" {
		found.Namespace = namespace
	}
	if session != nil {
		found.Session = *session
	}
	if user != "" {
		found.User = user
	}
	return found, nil
}

// DeleteContext will delete the named cluster
func (c *Config) DeleteContext(name string) error {
	if found := c.GetContextByName(name); found != nil {
		// should check if this context is in the current context
		if c.CurrentContext == name {
			return errors.New("cannot delete current context")
		}
		contexts := []*Context{}
		for _, context := range c.Contexts {
			if context.Name != found.Name {
				contexts = append(contexts, context)
			}
		}
		c.Contexts = contexts
		return nil
	}
	return errors.New("context not found")
}
