package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/nicle-lin/lillian/controller/api"
	"github.com/nicle-lin/lillian/controller/manager"
	"github.com/nicle-lin/lillian/helper/auth/builtin"
	"github.com/nicle-lin/lillian/version"
)

var (
	controllerManager *manager.Manager
)

func Server(c *cli.Context) {
	redisAddr := c.String("redisAddr")
	redisPassword := c.String("redisPassword")
	disableUsageInfo := c.Bool("disable-usage-info")
	listenAddr := c.String("listen")

	log.Infof("lillian CRM version: %s", version.Version)

	// default to builtin auth
	authenticator := builtin.NewAuthenticator("defaultlillian")

	controllerManager, err := manager.NewManager(redisAddr, redisPassword, disableUsageInfo, authenticator)
	if err != nil {
		log.Fatal(err)
	}

	apiConfig := api.ApiConfig{
		ListenAddr: listenAddr,
		Manager:    controllerManager,
	}

	lillianApi := api.NewApi(apiConfig)

	if err := lillianApi.Run(); err != nil {
		log.Fatal(err)
	}
}
