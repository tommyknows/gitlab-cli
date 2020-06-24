package config

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

// Config defines the structure of the config
// as a file.
type Config struct {
	Instances Instances           `json:"instances"`
	Contexts  map[string]*Context `json:"contexts"`

	CurrentContext string `json:"currentContext"`

	// PreferConfigContext indicates if the `currentContext` should take
	// precedence over the (filesystem-)local git repository.
	PreferConfigContext bool `json:"preferContext,omitempty"`

	name string // the file's name
	// if (filesystem-)local git repositories  should be ignored
	useConfigContext bool
}

type Instances map[string]*InstanceConfig

type InstanceConfig struct {
	Authentication *Authentication `json:"authentication,omitempty"`
	url            *url.URL
}

func (ic *InstanceConfig) apiURL() string {
	return "https://" + ic.url.Host + "/api/v4"
}

func (i *Instances) UnmarshalJSON(data []byte) error {
	type Alias Instances
	a := Alias(*i)
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	for iURL := range *i {
		u, err := url.Parse("https://" + iURL)
		if err != nil {
			return errors.Wrapf(err, "could not parse instance %v", iURL)
		}
		(*i)[iURL].url = u
	}
	return nil
}

type Authentication struct {
	Type                 AuthenticationType `json:"type"`
	*TokenAuthentication `json:",omitempty"`
	*BasicAuthentication `json:",omitempty"`
}

type AuthenticationType string

const (
	Token     AuthenticationType = "token"
	BasicAuth AuthenticationType = "basic-auth"
)

type Context struct {
	// the "default" group to use, a.k.a. the root of every operation.
	Group        string `json:"group,omitempty"`
	User         string `json:"user,omitempty"`
	InstanceName string `json:"instance"`

	// this is not populated at unmarshal because we cannot rely on
	// the order of unmarshaling.
	instanceConfig *InstanceConfig
}

type TokenAuthentication struct {
	Token string `json:"token"`
}

type BasicAuthentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Default() *Config {
	return &Config{
		Instances: make(map[string]*InstanceConfig),
		Contexts:  make(map[string]*Context),
	}
}

// Load tries to load a config from file. allocates a file if it should not exist
func Load(filename string, useConfigContext bool) (*Config, error) {
	if stat, _ := os.Stat(filename); stat == nil {
		defaultConfig := Default()
		defaultConfig.name = filename

		def, _ := json.Marshal(defaultConfig)
		return defaultConfig, ioutil.WriteFile(filename, def, 0600)
	}

	cont, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := Default()
	if err := json.Unmarshal(cont, c); err != nil {
		return nil, err
	}

	c.name = filename
	c.useConfigContext = useConfigContext

	return c, nil
}

func (c *Config) Write() error {
	cont, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.name, cont, 0600)
}
