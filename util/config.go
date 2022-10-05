package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PlexURL          string   `yaml:"plex_url"`
	SonarrURL        string   `yaml:"sonarr_url"`
	RadarrURL        string   `yaml:"radarr_url"`
	PlexToken        string   `yaml:"plex_token"`
	SonarrToken      string   `yaml:"sonarr_token"`
	RadarrToken      string   `yaml:"radarr_token"`
	ClientId         string   `yaml:"client_id"`
	DeleteMode       bool     `yaml:"delete_mode"`
	WatchedLibraries []string `yaml:"watched_libraries"`
}

func (c *Config) InitConfig() (*starr.Config, *starr.Config, error) {
	path := configDir()
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		id := uuid.New()
		c = &Config{
			PlexURL:          "http://127.0.0.1:32400",
			SonarrURL:        "http://127.0.0.1:8989",
			RadarrURL:        "http://127.0.0.1:7878",
			ClientId:         id.String(),
			DeleteMode:       true,
			WatchedLibraries: []string{"Movies", "TV Shows"},
		}
		c.WriteConfig()
		return nil, nil, errors.New("Created new config file! Please fill out all necessary fields!")
	}

	sonarr, radarr, err := c.ReadConfig()
	if err != nil {
		return nil, nil, err
	}
	c.PlexVerify()

	return sonarr, radarr, nil
}

func (c *Config) ReadConfig() (*starr.Config, *starr.Config, error) {
	path := configDir()
	ymldata, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err)
	}

	yaml.Unmarshal(ymldata, &c)

	timeout := time.Duration(1 * time.Second)
	services := map[string]string{c.PlexURL: "32400", c.SonarrURL: "8989", c.RadarrURL: "7878"}
	for k, v := range services {
		host, err := url.Parse(k)
		if err != nil {
			log.Fatal().Err(err).Msg("Malformed URL:")
		}

		port := host.Port()
		if port == "" {
			port = v
		}

		if _, err := net.DialTimeout("tcp", host.Hostname()+":"+port, timeout); k != "" && err != nil {
			msg := fmt.Sprintf("%v is down. Check if you entered the correct URL!", host.Hostname())
			return nil, nil, errors.New(msg)
		}
	}

	var sonarr, radarr *starr.Config
	if c.SonarrToken == "" && c.SonarrURL != "" {
		return nil, nil, errors.New("Sonarr API key missing!")
	} else {
		sonarr = starr.New(c.SonarrToken, c.SonarrURL, 0)
	}
	if c.RadarrToken == "" && c.RadarrURL != "" {
		return nil, nil, errors.New("Radarr API key missing!")
	} else {
		radarr = starr.New(c.RadarrToken, c.RadarrURL, 0)

	}

	return sonarr, radarr, nil
}

func (c *Config) WriteConfig() {
	path := configDir()
	base := strings.Split(path, "config.yml")

	ymldata, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatal().Err(err)
	}

	os.Mkdir(base[0], os.ModePerm)
	err = ioutil.WriteFile(path, ymldata, 0644)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func configDir() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err)
	}

	//TODO: Add os dependent paths
	path := fmt.Sprintf("%v/.config/clex/config.yml", user.HomeDir)
	return path
}
