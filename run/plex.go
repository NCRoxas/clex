package run

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/NCRoxas/clex/util"
	"github.com/rs/zerolog/log"
)

func ScanMedia(c *util.Config) ([]PlexMedia, []PlexMedia) {
	url := fmt.Sprintf("%v/library/sections/", c.Hosts.Plex)
	data := fetch(url, c.Tokens.Plex)

	// Check if library name exists on the server and add the correct key
	library := []Library{}
	for _, v := range data.Library {
		for _, w := range c.Libraries {
			if strings.EqualFold(v.Title, w) {
				library = append(library, v)
			}
		}
	}

	movies := make(chan PlexMedia)
	shows := make(chan PlexMedia)
	quit := make(chan struct{})

	movieList := []PlexMedia{}
	showList := []PlexMedia{}

	go filter(c, library, movies, shows, quit)

	for {
		select {
		case s := <-shows:
			showList = append(showList, s)
		case m := <-movies:
			movieList = append(movieList, m)
		case <-quit:
			return movieList, showList
		}
	}

}

// Filter through watched movies and shows
func filter(c *util.Config, library []Library, movies, shows chan PlexMedia, quit chan struct{}) {
	for _, v := range library {
		baseUrl := fmt.Sprintf("%v/library/sections/%v/all/", c.Hosts.Plex, v.Key)
		data := fetch(baseUrl+"?unwatched=0", c.Tokens.Plex)

		if v.Type == "movie" {
			for _, v := range data.Media {
				movies <- v
				log.Info().Str("Title", v.Title).Msg("Found watched movie:")
			}
		}

		if v.Type == "show" {
			for _, v := range data.Media {
				// Filter seasons containing watched episodes
				urlShow := fmt.Sprintf("%v%v?unwatched=0", c.Hosts.Plex, v.Key)
				dataSeason := fetch(urlShow, c.Tokens.Plex)

				for _, s := range dataSeason.Media {
					// Filter watched episodes of season
					urlEpisode := fmt.Sprintf("%v%v?unwatched=0", c.Hosts.Plex, s.Key)
					dataEpisode := fetch(urlEpisode, c.Tokens.Plex)

					for _, e := range dataEpisode.Media {
						//	Remove unfinished episodes
						if e.ViewCount > 0 && e.ViewOffset == 0 {
							shows <- e
							log.Info().
								Str("Show", e.GrandparentTitle).
								Str("Title", e.Title).
								Int64("Season", e.SeasonNumber).
								Int64("Episode", e.EpisodeNumber).
								Msg("Found watched show:")
						}
					}
				}
			}
		}
	}
	quit <- struct{}{}
}

// Fetches library data from the Plex server
func fetch(url string, token string) data {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal().Err(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", token)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err)
	}

	var c container
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&c); err != nil {
		log.Fatal().Err(err)
	}

	return c.Data
}
