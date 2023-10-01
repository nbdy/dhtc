package ui

import (
	"dhtc/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (c *Controller) DiscoverGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "discover", gin.H{
		"results": db.GetNRandomEntries(c.Database, 50),
		"path":    ctx.FullPath(),
	})
}

func (c *Controller) DiscoverPost(ctx *gin.Context) {
	N, err := strconv.Atoi(ctx.PostForm("limit"))
	if err != nil {
		N = 50
	}

	ctx.HTML(http.StatusOK, "discover", gin.H{
		"results": db.GetNRandomEntries(c.Database, N),
		"path":    ctx.FullPath(),
	})
}
