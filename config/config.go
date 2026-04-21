package config

import (
	"bufio"
	"os"
	"strings"
	"time"
)

const (
	defaultConfigDir      = "/.config/dp"
	defaultAliasesCache   = 1 * time.Hour
	defaultLocationsCache = 7 * 24 * time.Hour
	defaultRegionsCache   = 7 * 24 * time.Hour
)

type Config struct {
	configDir      string
	APIKey         string
	Output         string
	APIURL         string
	TestAPI        bool
	AliasesCache   time.Duration
	LocationsCache time.Duration
	RegionsCache   time.Duration
}

func (c *Config) Load() error {
	c.AliasesCache = defaultAliasesCache
	c.LocationsCache = defaultLocationsCache
	c.RegionsCache = defaultRegionsCache

	configPath, credsPath := c.configPaths()

	if configPath != "" {
		data, err := loadINI(configPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if data != nil {
			c.Output = data["output"]
			c.APIURL = data["api_url"]
			c.TestAPI = strings.EqualFold(data["test_api"], "true")

			if aliases := data["aliases_cache"]; aliases != "" {
				d, err := time.ParseDuration(aliases)
				if err != nil {
					return err
				}
				c.AliasesCache = d
			}

			if locations := data["locations_cache"]; locations != "" {
				d, err := time.ParseDuration(locations)
				if err != nil {
					return err
				}
				c.LocationsCache = d
			}

			if regions := data["regions_cache"]; regions != "" {
				d, err := time.ParseDuration(regions)
				if err != nil {
					return err
				}
				c.RegionsCache = d
			}
		}
	}

	if credsPath != "" {
		credsData, err := loadINI(credsPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if credsData != nil {
			if c.APIKey == "" {
				c.APIKey = credsData["api_key"]
			}
		}
	}

	return nil
}

func (c *Config) WithConfigDir(dir string) *Config {
	c.configDir = dir
	return c
}

func (c *Config) configPaths() (config, credentials string) {
	if c.configDir != "" {
		return c.configDir + "/config", c.configDir + "/credentials"
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", ""
	}
	return home + defaultConfigDir + "/config", home + defaultConfigDir + "/credentials"
}

func loadINI(path string) (keys map[string]string, err error) {
	f, openErr := os.Open(path) //nolint:gosec // path is from internal function
	if openErr != nil {
		return nil, openErr
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	keys = make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			keys[k] = v
		}
	}
	return keys, s.Err()
}
