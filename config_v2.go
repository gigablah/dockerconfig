package dockerconfig

import (
	"encoding/json"
	"io"
	"path/filepath"
)

const (
	configFileNameV2 = "config.json"
)

type configFileV2 struct {
	*ConfigFile
	ConfigReadWriter `json:"-"`
}

func (c *configFileV2) ConfigDir() string {
	configDir := c.configDir
	if configDir == "" {
		configDir = filepath.Join(getHomeDir(), ".docker")
	}
	return configDir
}

func (c *configFileV2) Filename() string {
	filename := c.filename
	if filename == "" {
		filename = configFileNameV2
	}
	return filepath.Join(c.ConfigDir(), filename)
}

func (c *configFileV2) LoadFromReader(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}
	var err error
	for addr, ac := range c.AuthConfigs {
		ac.Username, ac.Password, err = DecodeAuth(ac.Auth)
		if err != nil {
			return err
		}
		ac.Auth = ""
		ac.ServerAddress = addr
		c.AuthConfigs[addr] = ac
	}
	return nil
}

func (c *configFileV2) SaveToWriter(w io.Writer) error {
	// Encode sensitive data into a new/temp struct
	tmpAuthConfigs := make(map[string]AuthConfig, len(c.AuthConfigs))
	for k, authConfig := range c.AuthConfigs {
		authCopy := authConfig
		// encode and save the authstring, while blanking out the original fields
		authCopy.Auth = EncodeAuth(&authCopy)
		authCopy.Username = ""
		authCopy.Password = ""
		authCopy.ServerAddress = ""
		tmpAuthConfigs[k] = authCopy
	}

	saveAuthConfigs := c.AuthConfigs
	c.AuthConfigs = tmpAuthConfigs
	defer func() { c.AuthConfigs = saveAuthConfigs }()

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
