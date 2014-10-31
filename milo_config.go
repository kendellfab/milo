package milo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Simple milo config object.
type Config struct {
	Bind              string                 `json:"bind"`       // Which interface to bind to.  0.0.0.0
	Port              int                    `json:"port"`       // What port number to bind to
	PortIncrement     bool                   `json:"port_inc"`   // If the given port is used, increment one and try again.  Good for supervisord deployments
	TemplateDirectory string                 `json:"tpl_dir"`    // Where the templates are found
	CacheTemplates    bool                   `json:"cache_tpls"` // Flag to handle template caching
	AssetDirectory    string                 `json:"asset_dir"`  // Where the assets are found
	SessionKeys       []string               `json:"sess_keys"`  // Session keys used by gorilla sessions
	CatchAll          bool                   `json:"catch_all"`  // A last registered route to handle assets in the asset directory i.e. robots.txt
	AppConfig         map[string]interface{} `json:"app_config"` // A set of application configs i.e. site name, environment variables
}

// Get the connection string from the config object.
func (c *Config) GetConnectionString() string {
	return fmt.Sprintf("%s:%d", c.Bind, c.Port)
}

// Get an app config item as a string.
func (c *Config) GetConfigString(key string) string {
	if val, ok := c.AppConfig[key]; ok {
		if item, asStr := val.(string); asStr {
			return item
		}
	}
	return ""
}

// Get an app config item as an int.
func (c *Config) GetConfigInt(key string) int {
	if val, ok := c.AppConfig[key]; ok {
		if item, asInt := val.(int); asInt {
			return item
		}
	}
	return -1
}

// Get an app config as another map.
func (c *Config) GetConfigMap(key string) map[string]interface{} {
	if val, ok := c.AppConfig[key]; ok {
		if item, asMap := val.(map[string]interface{}); asMap {
			return item
		}
	}
	return nil
}

// Build the config from file from a json serialization.
func ConfigFromFile(path string) *Config {
	byteArr, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	var config Config
	if pErr := json.Unmarshal(byteArr, &config); pErr != nil {
		fmt.Println(pErr)
		os.Exit(2)
	}
	return &config
}

// Get a default config struct with some sane defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:              7000,
		TemplateDirectory: "tpls",
		AssetDirectory:    "static",
	}
}
