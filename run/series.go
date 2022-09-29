package run

import (
	"strings"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

func QueueSeries(sc *starr.Config, media []PlexMedia) {
	s := sonarr.New(sc)
	series, _ := s.GetAllSeries()

	watchedShows := map[int64]map[int64][]int64{}
	for _, show := range series {
		episodes := []int64{}
		season := map[int64][]int64{}

		for _, m := range media {
			if strings.EqualFold(show.Title, m.GrandparentTitle) {
				episodes = append(episodes, m.EpisodeNumber)
				season[m.SeasonNumber] = episodes

				watchedShows[show.ID] = season
			}
		}
	}

	// Get episode file numbers
	marked := Marked{}
	for id, info := range watchedShows {
		sonarrFileInfo, _ := s.GetSeriesEpisodes(id)

		for season, episodes := range info {
			for _, file := range sonarrFileInfo {
				if file.SeasonNumber == season && findEpisode(episodes, file.EpisodeNumber) {
					marked.Watched = append(marked.Watched, file.ID)                      // ID of episode
					marked.EpisodeFiles = append(marked.EpisodeFiles, file.EpisodeFileID) // ID of file on disk
				}
			}
		}
	}

	// Unmonitor episodes
	s.MonitorEpisode(marked.Watched, false)

	// Unmonitor seasons
	for id, _ := range watchedShows {
		show, _ := s.GetSeriesByID(id)
		episodes, _ := s.GetSeriesEpisodes(id)
		watchedSe := 0

		for i := range show.Seasons {
			watchedEp := 0
			total := show.Seasons[i].Statistics.TotalEpisodeCount

			// Count unmonitored episodes
			for _, ep := range episodes {
				if ep.SeasonNumber == int64(i) && ep.Monitored == false {
					watchedEp++
				}
			}

			// Unmonitor seasons
			if watchedEp == total && show.Seasons[i].Monitored == true {
				log.Info().Str("Title", show.Title).Int("Season", i).Msg("Unmonitoring season")
				show.Seasons[i].Monitored = false
			}

			if show.Seasons[i].Monitored == false {
				watchedSe++
			}
			watchedEp = 0
		}

		// Unmonitor show
		if show.Ended && watchedSe == show.Statistics.SeasonCount+1 {
			log.Info().Str("Title", show.Title).Msg("Unmonitoring show")
			show.Monitored = false
		}

		if err := s.UpdateSeries(id, show); err != nil {
			log.Error().Msg(err.Error())
		}
	}

	// Delete Episodefiles
	for _, file := range marked.EpisodeFiles {
		s.DeleteEpisodeFile(file)
	}
}

func findEpisode(slice []int64, val int64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
