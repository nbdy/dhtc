package ui

import (
	"dhtc/config"
	"github.com/gin-gonic/gin"
	"github.com/ostafen/clover/v2"
)

type Controller struct {
	Database      *clover.DB
	Configuration *config.Configuration
}

func RunWebServer(configuration *config.Configuration, database *clover.DB) {
	srv := gin.Default()

	uiCtrl := Controller{
		Database:      database,
		Configuration: configuration,
	}

	srv.GET("", uiCtrl.Dashboard)
	srv.GET("/dashboard", uiCtrl.Dashboard)
	srv.GET("/search", uiCtrl.SearchGet)
	srv.POST("/search", uiCtrl.SearchPost)
	srv.GET("/discover", uiCtrl.DiscoverGet)
	srv.POST("/discover", uiCtrl.DiscoverPost)
	srv.GET("/watches", uiCtrl.WatchGet)
	srv.POST("/watches", uiCtrl.WatchPost)
	srv.GET("/blacklist", uiCtrl.BlacklistGet)
	srv.POST("/blacklist", uiCtrl.BlacklistPost)

	srv.Static("/css", "static/css")
	srv.Static("/js", "static/js")

	err := srv.Run(configuration.Address)

	if err != nil {
		return
	}
}
