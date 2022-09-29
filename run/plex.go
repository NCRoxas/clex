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
	url := fmt.Sprintf("%v/library/sections/", c.PlexURL)
	data := fetch(url, c.PlexToken)

	watchedLibs := []Library{}
	for _, v := range data.Library {
		for _, w := range c.WatchedLibraries {
			if strings.EqualFold(v.Title, w) {
				watchedLibs = append(watchedLibs, v)
			}
		}
	}

	watchedMovies := []PlexMedia{}
	watchedSeries := []PlexMedia{}
	for _, v := range watchedLibs {
		filterMedia := getWatchedMedia(c, v.Key)
		if v.Type == "movie" {
			watchedMovies = append(watchedMovies, filterMedia...)
		}

		if v.Type == "show" {
			watchedSeries = append(watchedSeries, filterMedia...)
		}

	}

	return watchedMovies, watchedSeries
}

func getWatchedMedia(c *util.Config, key string) []PlexMedia {
	baseUrl := fmt.Sprintf("%v/library/sections/%v/all/", c.PlexURL, key)
	dataShows := fetch(baseUrl, c.PlexToken)
	dataMovies := fetch(baseUrl+"?unwatched=0", c.PlexToken)

	watched := []PlexMedia{}
	for _, v := range dataMovies.Media {
		if v.Type == "movie" {
			watched = append(watched, v)
			log.Info().Str("Title", v.Title).Msg("Found watched movie:")
		}
	}

	for _, v := range dataShows.Media {
		if v.Type == "show" {
			// Filter seasons containing watched episodes
			urlShow := fmt.Sprintf("%v%v?unwatched=0", c.PlexURL, v.Key)
			dataSeason := fetch(urlShow, c.PlexToken)

			for _, s := range dataSeason.Media {
				// Filter watched episodes of season
				urlEpisode := fmt.Sprintf("%v%v?unwatched=0", c.PlexURL, s.Key)
				dataEpisode := fetch(urlEpisode, c.PlexToken)

				for _, e := range dataEpisode.Media {
					//	Remove unfinished episodes
					if e.ViewCount > 0 && e.ViewOffset == 0 {
						watched = append(watched, e)
						log.Info().Str("Show", e.GrandparentTitle).Str("Title", e.Title).Int64("Season", e.SeasonNumber).Int64("Episode", e.EpisodeNumber).Msg("Found watched show:")
					}
				}
			}
		}
	}

	return watched
}

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
