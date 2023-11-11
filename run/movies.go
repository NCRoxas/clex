package run

import (
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

	request := &radarr.BulkEdit{
		MovieIDs:           queue.Watched,
		DeleteFiles:        starr.False(),
		Monitored:          starr.False(),
		AddImportExclusion: starr.False(),
	}
	if c.Delete {
		request.DeleteFiles = starr.True()
	}
	if c.Exclude {
		request.AddImportExclusion = starr.True()
	}
	r.EditMovies(request)
	log.Info().Msgf("Edited %d movies", len(queue.Watched))
}
