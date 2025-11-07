package ui

import (
	"dhtc/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Dashboard(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "dashboard", gin.H{
		"info_hash_count": db.GetInfoHashCount(c.Database),
		"path":            ctx.FullPath(),
		"statistics":      c.Configuration.Statistics,
	})
}
