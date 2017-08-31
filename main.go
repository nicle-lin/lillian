package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/nicle-lin/lillian/controller/server"
	"github.com/nicle-lin/lillian/version"
	"os"
)

const STORE_KEY = "lillian"

func main() {
	app := cli.NewApp()
	app.Name = "lillian"
	app.Usage = "lillian crm"
	app.Version = version.Version + "(" + version.GitCommit + ")"
	app.Author = "nicle-lin"
	app.Email = "dghpgyss@163.com"
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:   "server",
			Usage:  "run lillian controller",
			Action: server.Server,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "listen,l",
					Usage: "listen address",
					Value: ":5525",
				},
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,D",
			Usage: "enable debug",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
