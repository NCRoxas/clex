package run

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golift.io/starr"
	"golift.io/starr/radarr"
)

// TODO: Refactor some stuff
func QueueMovies(sc *starr.Config, media []PlexMedia, deleteMode bool) {
	r := radarr.New(sc)
	movies, _ := r.GetMovie(0)

	marked := Marked{}
	for _, movie := range movies {
		for _, w := range media {
			if strings.EqualFold(movie.Title, w.Title) {
				marked.Watched = append(marked.Watched, movie.ID)
			}
		}
	}

	// Delete movie files and add exclusion
	if deleteMode {
		url := fmt.Sprintf("%v/api/v3/movie/editor?apikey=%v", sc.URL, sc.APIKey)
		request, _ := json.Marshal(map[string]any{
			"movieIds":           marked.Watched,
			"monitored":          false,
			"deleteFiles":        true,
			"addImportExclusion": true,
		})
		deleteMovie(url, request)
	}
}

func deleteMovie(url string, request []byte) {
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(request))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	if _, err := client.Do(req); err != nil {
		log.Fatalln(err)
	}
}
