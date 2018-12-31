package main

import (
	"os"
	"syscall"

	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
)

var osChan = make(chan os.Signal)
var fs = afero.NewOsFs()
var exiter = log.Exit

func main() {
	log.SetLevel(log.InfoLevel)

	c := NewCron(fs)
	if err := c.LoadConfig("/config/config.json"); err != nil {
		log.Errorln(err)
		exiter(1)
	}

	if err := c.Run(osChan, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Errorln(err)
		exiter(1)
	}
}
