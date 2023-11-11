package util

import (
	"errors"
	"fmt"
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
		c.onboard()
	}

	return c.readConfig()
}

func (c *Config) readConfig() (*starr.Config, *starr.Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
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

		if _, err := net.DialTimeout("tcp", host.Hostname()+":"+port, time.Duration(10*time.Second)); service != "" &&
			err != nil {
			msg := fmt.Sprintf("%v is down. Check if you entered the correct URL!", host.Hostname())
			return nil, nil, errors.New(msg)
		}
	}

	var sonarr, radarr *starr.Config
	if c.Tokens.Sonarr == "" && c.Hosts.Sonarr != "" {
		log.Warn().Msg("Set your Sonarr API key: ")
		fmt.Scanf("%s", &c.Tokens.Sonarr)
		c.WriteConfig()
		if c.Tokens.Sonarr == "" {
			return nil, nil, errors.New("sonarr token is missing")
		}
	} else if c.Tokens.Sonarr != "" && c.Hosts.Sonarr != "" {
		sonarr = starr.New(c.Tokens.Sonarr, c.Hosts.Sonarr, 0)
	}
	if c.Tokens.Radarr == "" && c.Hosts.Radarr != "" {
		log.Warn().Msg("Set your Radarr API key: ")
		fmt.Scanf("%s", &c.Tokens.Radarr)
		c.WriteConfig()
		if c.Tokens.Radarr == "" {
			return nil, nil, errors.New("radarr token is missing")
		}
	} else if c.Tokens.Radarr != "" && c.Hosts.Radarr != "" {
		radarr = starr.New(c.Tokens.Radarr, c.Hosts.Radarr, 0)
	}
	if c.Hosts.Plex == "" {
		log.Warn().Msg("Set your Plex host: ")
		fmt.Scanf("%s", &c.Hosts.Plex)
		c.WriteConfig()
		if c.Hosts.Plex == "" {
			return nil, nil, errors.New("plex host is missing")
		}
		_, err := url.ParseRequestURI(c.Hosts.Plex)
		if err != nil {
			return nil, nil, errors.New("seems like your plex host url is malformed")
		}

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
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Info().Msgf("Your config file has been updated here: %v\n", configPath)
}

func parseBooleanInput(input string) (bool, error) {
	input = strings.ToLower(input)
	switch input {
	case "yes":
		return true, nil
	case "no":
		return false, nil
	case "y":
		return true, nil
	case "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean input: %s", input)
	}
}

func (c *Config) onboard() {
	clientId := uuid.New() // Generate random ID on first use
	c = &Config{
		Delete:    true,
		Exclude:   true,
		ClientId:  clientId.String(),
		Libraries: []string{"Movies", "TV Shows"},
	}

	// Required plex fields
	log.Info().Msg("New installation detected! Generating a new config...")
	log.Info().Msgf("Your new client ID: %v", clientId.String())
	log.Info().Msg("Set your Plex host: (e.g. http://localhost:32400)")
	fmt.Print("> ")
	fmt.Scanf("%s", &c.Hosts.Plex)
	_, err := url.ParseRequestURI(c.Hosts.Plex)
	if err != nil {
		log.Fatal().Err(err).Msg("Seems like your Plex host url is malformed:")
	}

	// Optional arr's
	log.Info().Msg("Set your Sonarr host: (e.g. http://localhost:8989)")
	fmt.Print("> ")
	fmt.Scanf("%s", &c.Hosts.Sonarr)
	if c.Hosts.Sonarr != "" {
		_, err = url.ParseRequestURI(c.Hosts.Sonarr)
		if err != nil {
			log.Fatal().Err(err).Msg("Seems like your Sonarr host url is malformed:")
		}
	}
	log.Info().Msg("Set your Radarr host: (e.g. http://localhost:8989)")
	fmt.Print("> ")
	fmt.Scanf("%s", &c.Hosts.Radarr)
	if c.Hosts.Radarr != "" {
		_, err = url.ParseRequestURI(c.Hosts.Radarr)
		if err != nil {
			log.Fatal().Err(err).Msg("Seems like your Radarr host url is malformed:")
		}
	}

	// If user inputs host token is needed
	if c.Hosts.Sonarr != "" {
		log.Info().Msg("Set your Sonarr API token: ")
		fmt.Print("> ")
		fmt.Scanf("%s", &c.Tokens.Sonarr)
		if c.Tokens.Sonarr == "" {
			log.Fatal().Msg("Sonarr API key is missing!")
		}
	}
	if c.Hosts.Radarr != "" {
		log.Info().Msg("Set your Radarr API token: ")
		fmt.Print("> ")
		fmt.Scanf("%s", &c.Tokens.Radarr)
		if c.Tokens.Radarr == "" {
			log.Fatal().Msg("Radarr API key is missing!")
		}
	}

	// Optional fields with defaults
	var deleteMode, excludeMode string
	log.Info().Msg("Do you want to delete shows/movies after watching <yes/no>? Default: yes")
	fmt.Print("> ")
	fmt.Scanf("%s", &deleteMode)
	if deleteMode != "" {
		c.Delete, _ = parseBooleanInput(deleteMode)
	}
	log.Info().
		Msg("Do you want to add shows/movies to the exclusion list after watching <yes/no>? Default: yes")
	fmt.Print("> ")
	fmt.Scanf("%s", &excludeMode)
	if excludeMode != "" {
		c.Exclude, _ = parseBooleanInput(excludeMode)
	}

	var libraries string
	log.Info().
		Msg("Which libraries do you want to watch? Default: Movies,TV Shows")
	fmt.Print("> ")
	fmt.Scanln(&libraries)
	if libraries != "" {
		c.Libraries = strings.Split(libraries, ",")
	}

	c.WriteConfig()
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
