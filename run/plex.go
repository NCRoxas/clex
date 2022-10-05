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

	movies := make(chan PlexMedia)
	shows := make(chan PlexMedia)
	quit := make(chan int)

	movieList := []PlexMedia{}
	showList := []PlexMedia{}

	go func() {
		for _, v := range watchedLibs {
			filterMedia(c, v.Key, v.Type, movies, shows)
		}
		quit <- 0
	}()

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

func filterMedia(c *util.Config, key, mediaType string, movies, shows chan PlexMedia) {
	baseUrl := fmt.Sprintf("%v/library/sections/%v/all/", c.PlexURL, key)

	if mediaType == "movie" {
		data := fetch(baseUrl+"?unwatched=0", c.PlexToken)
		//watchedMovies := []PlexMedia{}

		for _, v := range data.Media {
			//watchedMovies = append(watchedMovies, v)
			movies <- v
			log.Info().Str("Title", v.Title).Msg("Found watched movie:")
		}
	}

	if mediaType == "show" {
		data := fetch(baseUrl, c.PlexToken)
		//watchedSeries := []PlexMedia{}

		for _, v := range data.Media {
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
						//watchedSeries = append(watchedSeries, e)
						shows <- e
						log.Info().Str("Show", e.GrandparentTitle).Str("Title", e.Title).Int64("Season", e.SeasonNumber).Int64("Episode", e.EpisodeNumber).Msg("Found watched show:")
					}
				}
			}
		}
		//shows <- watchedSeries
	}
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
