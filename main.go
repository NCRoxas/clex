package main

import (
	"flag"
	"os"
	"time"

	"github.com/NCRoxas/clex/run"
	"github.com/NCRoxas/clex/util"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	s := gocron.NewScheduler(time.UTC)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	interval := flag.String(
		"interval",
		"once",
		"Set interval of cleanup job. Choices are: once, daily, weekly, monthly",
	)
	runAt := flag.String("time", "05:00", "Set time when to run cleanup job")
	flag.Parse()

	if *interval == "daily" {
		log.Info().Str("Time", *runAt).Msg("Running daily cleanup job at")
		s.Every(1).Day().At(*runAt).Do(job)
	}
	if *interval == "weekly" {
		log.Info().Str("Time", *runAt).Str("Day", "Monday").Msg("Running weekly cleanup job at")
		s.Every(1).Week().Monday().At(*runAt).Do(job)
	}
	if *interval == "monthly" {
		log.Info().Str("Time", *runAt).Str("Day", "Monday").Msg("Running monthly cleanup job at")
		s.MonthFirstWeekday(time.Monday).At(*runAt).Do(job)
	}
	if *interval == "once" {
		job()
	} else {
		s.StartBlocking()
	}
}

func job() {
	var c util.Config
	sonarr, radarr, err := c.InitConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config")
	}
	c.PlexVerify()
	watchedMovies, watchedSeries := run.ScanMedia(&c)
	if radarr != nil {
		run.QueueMovies(radarr, watchedMovies, &c)
	} else {
		log.Info().Msg("Radarr is not enabled. Skipping...")
	}
	if sonarr != nil {
		run.QueueSeries(sonarr, watchedSeries, &c)
	} else {
		log.Info().Msg("Sonarr is not enabled. Skipping...")
	}
}
