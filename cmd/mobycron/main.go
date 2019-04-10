package main

import (
	"io"
	"os"
	"syscall"

	"github.com/pfillion/mobycron/pkg/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var osChan = make(chan os.Signal)
var fs = afero.NewOsFs()
var exiter = log.Exit
var output io.Writer = os.Stdout

func main() {
	log.SetOutput(output)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})

	c := cron.NewCron(fs)
	if err := c.Run(osChan, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Errorln(err)
		exiter(1)
	}
}
