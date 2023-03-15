package run

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/NCRoxas/clex/util"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

func QueueMovies(sc *starr.Config, media []PlexMedia, c *util.Config) {
	r := radarr.New(sc)
	movies, _ := r.GetMovie(0)

	queue := Queue{}
	for _, movie := range movies {
		for _, watched := range media {
			if strings.EqualFold(movie.Title, watched.Title) {
				queue.Watched = append(queue.Watched, movie.ID)
			}
		}
	}

	url := fmt.Sprintf("%v/api/v3/movie/editor?apikey=% v", sc.URL, sc.APIKey)
	request := map[string]any{
		"movieIds":           queue.Watched,
		"monitored":          false,
		"deleteFiles":        false,
		"addImportExclusion": false,
	}

	if c.Delete {
		//editMovie(url, request)
		// r.DeleteMovies(&radarr.BulkEdit{
		// 	MovieIDs:           queue.Watched,
		// 	DeleteFiles:        starr.True(),
		// 	Monitored:          starr.False(),
		// 	AddImportExclusion: starr.True(),
		// })
		request["deleteFiles"] = true
		log.Info().Msgf("Delete enabled")
	}
	if c.Exclude {
		request["addImportExclusion"] = true
		log.Info().Msgf("Exclusion enabled")
	}
	req, _ := json.Marshal(request)
	editMovie(url, req)
	log.Info().Msgf("Edited %d movies", len(queue.Watched))

}

// TODO: Refactor
func editMovie(url string, request []byte) {
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(request))
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	if _, err := client.Do(req); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
