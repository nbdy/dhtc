package ui

import (
	"dhtc/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SearchGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "search", gin.H{
		"path": ctx.FullPath(),
	})
}

func (c *Controller) SearchPost(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "search", gin.H{
		"results": db.FindBy(
			c.Database,
			ctx.PostForm("key"),
			ctx.PostForm("match-type"),
			ctx.PostForm("search-input")),
		"path": ctx.FullPath(),
	})
}
