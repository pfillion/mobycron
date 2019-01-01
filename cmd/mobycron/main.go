package main

import (
	"os"
	"syscall"

	"github.com/pfillion/mobycron/pkg/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var osChan = make(chan os.Signal)
var fs = afero.NewOsFs()
var exiter = log.Exit

func main() {
	log.SetLevel(log.InfoLevel)

	c := cron.NewCron(fs)
	if err := c.LoadConfig("/configs/config.json"); err != nil {
		log.Errorln(err)
		exiter(1)
	}

	if err := c.Run(osChan, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Errorln(err)
		exiter(1)
	}
}
