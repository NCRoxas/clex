package util

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type PinData struct {
	ID               int       `json:"id"`
	Code             string    `json:"code"`
	Product          string    `json:"product"`
	Trusted          bool      `json:"trusted"`
	ClientIdentifier string    `json:"clientIdentifier"`
	ExpiresIn        int       `json:"expiresIn"`
	CreatedAt        time.Time `json:"createdAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
	AuthToken        string    `json:"authToken"`
	NewRegistration  bool      `json:"newRegistration"`
}

const (
	plexPin  = "https://plex.tv/api/v2/pins/"
	plexUser = "https://plex.tv/api/v2/user/"
)

func (c *Config) PlexVerify() {
	host, err := url.Parse(c.PlexURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Malformed URL:")
		return
	}
	port := host.Port()
	if port == "" {
		port = "32400"
	}

	timeout := time.Duration(1 * time.Second)
	if _, err := net.DialTimeout("tcp", host.Hostname()+":"+port, timeout); err != nil {
		log.Fatal().Err(err).Msg("Host is down. Check if you entered the correct URL of your Plex server!")
		return
	}

	if c.PlexToken != "" {
		c.CheckToken()
	} else {
		log.Info().Msg("Requesting new token...")
		pin, err := c.GeneratePin()
		if err != nil {
			log.Fatal().Err(err)
		}

		pin.AuthUrl()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go pin.Poll(&wg)
		wg.Wait()

		c.PlexToken = pin.AuthToken
		c.WriteConfig()
	}
}

func (c *Config) GeneratePin() (*PinData, error) {
	form := url.Values{
		"strong":                   {"true"},
		"X-Plex-product":           {"Clex"},
		"X-Plex-client-Identifier": {c.ClientId},
	}
	req, err := http.NewRequest("POST", plexPin, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err)
	}

	defer resp.Body.Close()

	var p PinData
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *PinData) AuthUrl() {
	v := url.Values{
		"clientID":                 {p.ClientIdentifier},
		"code":                     {p.Code},
		"context[device][product]": {p.Product},
	}
	url := fmt.Sprintf("https://app.plex.tv/auth#?%s", v.Encode())
	log.Info().Str("URL", url).Msg("Authentication URL:")
}

func (c *Config) CheckToken() {
	form := url.Values{
		"strong":                   {"true"},
		"X-Plex-product":           {"Autoclean Plex"},
		"X-Plex-client-Identifier": {c.ClientId},
		"X-Plex-Token":             {c.PlexToken},
	}
	req, err := http.NewRequest("GET", plexUser, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal().Err(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err)
	}

	// Redo Authentication Process
	if resp.StatusCode == 401 {
		c.PlexToken = ""
		c.WriteConfig()
		c.PlexVerify()
	}
}

func (p *PinData) Poll(wg *sync.WaitGroup) {
	status := true
	for status {
		v := url.Values{
			"code":                     {p.Code},
			"X-Plex-client-Identifier": {p.ClientIdentifier},
		}
		req, err := http.NewRequest("GET", plexPin+strconv.Itoa(p.ID), strings.NewReader(v.Encode()))
		if err != nil {
			log.Fatal().Err(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal().Err(err)
		}

		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&p); err != nil {
			log.Fatal().Err(err)
		}

		if p.AuthToken != "" {
			status = false
		}
		time.Sleep(time.Second * 1)
	}
	wg.Done()
}
