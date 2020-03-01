package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pfillion/mobycron/pkg/cron"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Cronner keeps track of any number of jobs, invoking the associated Job as
// specified by the schedule. It may be started and stopped.
type Cronner interface {
	LoadConfig(filename string) error
	Start()
	Stop() context.Context
}

// Handler scan and listen docker messages of containers labeled for crontab
type Handler interface {
	Scan() error
	ListenContainer()
	ListenService()
}

var (
	osChan  chan os.Signal
	handler Handler
	cronner Cronner
	cmdRoot *cli.App
	cfg     = config{}
)

type config struct {
	cfgFile     string
	dockerMode  bool
	parseSecond bool
}

func initApp(ctx *cli.Context) error {
	c := cron.NewCron(cfg.parseSecond)
	if cfg.dockerMode {
		h, err := cron.NewHandler(c)
		if err != nil {
			return err
		}
		handler = h
	} else {
		handler = nil
	}

	cronner = c
	osChan = make(chan os.Signal)

	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})
	return nil
}

func startApp(ctx *cli.Context) error {
	sig := []os.Signal{syscall.SIGINT, syscall.SIGTERM}

	if cfg.cfgFile != "" {
		if err := cronner.LoadConfig(cfg.cfgFile); err != nil {
			return err
		}
	}

	if cfg.dockerMode {
		if err := handler.Scan(); err != nil {
			return err
		}

		handler.ListenContainer()
		handler.ListenService()
	}

	cronner.Start()

	log.WithFields(log.Fields{
		"func":   "main.startApp",
		"signal": sig,
	}).Infoln("cron is running and waiting signal for stop")

	signal.Notify(osChan, sig...)
	<-osChan

	cronner.Stop()
	// TODO: Refactoring of all log. Check if useful and complete. Think if it possible to have class for manage logging OR methods to make all fields correctly
	// TODO: Refactoring of all test for check log with Fields like handler_test working with output but with field and value
	// TODO: Refactoring of all log Fields to manage sub object ex: event.ID event.Actor.ID. It will be ready for kibana and elasticsearch
	// TODO: Migrate to urfave/cli/v2
	// TODO: Refactoring all tests for verify all fields logged in the main test case
	// TODO: change label action to be 'start' by default
	// TODO: evaluate if we need to manage 'Exec' action when the container stop or die

	return nil
}

func main() {
	err := cmdRoot.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	cmdRoot = cli.NewApp()
	cmdRoot.Before = initApp
	cmdRoot.Action = startApp

	// Global options
	cmdRoot.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name:        "docker-mode, d",
			EnvVar:      "MOBYCRON_DOCKER_MODE",
			Destination: &cfg.dockerMode,
			Usage:       "activate docker mode (default: true)",
		},
		cli.BoolFlag{
			Name:        "parse-second, s",
			EnvVar:      "MOBYCRON_PARSE_SECOND",
			Destination: &cfg.parseSecond,
			Usage:       "accept an optional seconds field at the beginning of the cron spec (default: false)",
		},
		cli.StringFlag{
			Name:        "config-file, f",
			EnvVar:      "MOBYCRON_CONFIG_FILE",
			Destination: &cfg.cfgFile,
			Usage:       "set file path to schedule all job like a crontab file",
		},
	}
}
