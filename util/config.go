package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Hosts struct {
		Plex   string `yaml:"plex"`
		Sonarr string `yaml:"sonarr"`
		Radarr string `yaml:"radarr"`
	} `yaml:"hosts"`
	Tokens struct {
		Plex   string `yaml:"plex"`
		Sonarr string `yaml:"sonarr"`
		Radarr string `yaml:"radarr"`
	} `yaml:"tokens"`
	Delete    bool     `yaml:"delete"`
	Exclude   bool     `yaml:"exclude"`
	ClientId  string   `yaml:"client_id"`
	Libraries []string `yaml:"libraries"`
}

func (c *Config) InitConfig() (*starr.Config, *starr.Config, error) {
	configPath := getConfigPath()
	_, err := os.Stat(configPath)

	if errors.Is(err, os.ErrNotExist) {
		clientId := uuid.New() // Generate random ID on first use
		c = &Config{
			Delete:    true,
			Exclude:   true,
			ClientId:  clientId.String(),
			Libraries: []string{"Movies", "TV Shows"},
		}
		c.WriteConfig()
		return nil, nil, errors.New("Created new config file! Please fill out all necessary fields!")
	}

	sonarr, radarr, err := c.ReadConfig()
	return sonarr, radarr, err
}

func (c *Config) ReadConfig() (*starr.Config, *starr.Config, error) {
	configPath := getConfigPath()
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal().Err(err)
	}

	yaml.Unmarshal(data, &c)
	services := []string{
		c.Hosts.Plex,
		c.Hosts.Sonarr,
		c.Hosts.Radarr,
	}

	for _, service := range services {
		host, err := url.Parse(service)
		if err != nil {
			log.Fatal().Err(err).Msg("Malformed URL:")
		}

		port := host.Port()
		if port == "" && host.Scheme == "http" {
			port = "80"
		} else if port == "" && host.Scheme == "https" {
			port = "443"
		}

		if _, err := net.DialTimeout("tcp", host.Hostname()+":"+port, time.Duration(10*time.Second)); service != "" && err != nil {
			msg := fmt.Sprintf("%v is down. Check if you entered the correct URL!", host.Hostname())
			return nil, nil, errors.New(msg)
		}
	}

	var sonarr, radarr *starr.Config
	if c.Tokens.Sonarr == "" && c.Hosts.Sonarr != "" {
		return nil, nil, errors.New("Sonarr API key is missing!")
	} else if c.Tokens.Sonarr != "" && c.Hosts.Sonarr != "" {
		sonarr = starr.New(c.Tokens.Sonarr, c.Hosts.Sonarr, 0)
	}
	if c.Tokens.Radarr == "" && c.Hosts.Radarr != "" {
		return nil, nil, errors.New("Radarr API key is missing!")
	} else if c.Tokens.Radarr != "" && c.Hosts.Radarr != "" {
		radarr = starr.New(c.Tokens.Radarr, c.Hosts.Radarr, 0)
	}
	if c.Hosts.Plex == "" {
		return nil, nil, errors.New("Please add your plex host!")
	}
	return sonarr, radarr, nil
}

func (c *Config) WriteConfig() {
	configPath := getConfigPath()
	base := strings.Split(configPath, "config.yml")

	data, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatal().Err(err)
	}

	os.Mkdir(base[0], os.ModePerm)
	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func getConfigPath() string {
	if runtime.GOOS != "linux" {
		log.Fatal().Msg("Only Linux is supported at the moment!")
	}

	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err)
	}

	return fmt.Sprintf("%v/.config/clex/config.yml", user.HomeDir)
}
