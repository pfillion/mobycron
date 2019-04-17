package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pfillion/mobycron/pkg/cron"
	"github.com/pfillion/mobycron/pkg/events"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Cron keeps track of any number of entries, invoking the associated Job as
// specified by the schedule. It may be started and stopped.
type Cron interface {
	LoadConfig(filename string) error
	Start()
	Stop()
}

// Handler scan and listen docker messages of containers labeled for crontab
type Handler interface {
	Scan() error
	Listen() error
}

type cronApp struct {
	cron    Cron
	osChan  chan os.Signal
	handler Handler
}

var before = initApp
var action = startApp
var app cronApp

func initApp(ctx *cli.Context) error {
	cron := cron.NewCron()
	osChan := make(chan os.Signal)
	handler, err := events.NewHandler(cron)
	if err != nil {
		return err
	}

	app = cronApp{
		cron:    cron,
		osChan:  osChan,
		handler: handler,
	}

	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})
	return nil
}

func startApp(ctx *cli.Context) error {
	sig := []os.Signal{syscall.SIGINT, syscall.SIGTERM}

	if err := app.cron.LoadConfig("/configs/config.json"); err != nil {
		return err
	}

	// if err := app.handler.Scan(); err != nil {
	// 	return err
	// }

	// if err := app.handler.Listen(); err != nil {
	// 	return err
	// }

	app.cron.Start()

	log.WithFields(log.Fields{
		"func":   "Run",
		"signal": sig,
	}).Infoln("cron is running and waiting signal for stop")

	signal.Notify(app.osChan, sig...)
	<-app.osChan

	app.cron.Stop()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Before = before
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
