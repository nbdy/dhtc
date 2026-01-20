package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) WatchGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "watches", gin.H{
		"path":    ctx.FullPath(),
		"results": c.Database.GetWatchEntries(),
	})
}

func (c *Controller) WatchPost(ctx *gin.Context) {
	opOk := false
	op := ctx.PostForm("op")
	if op == "add" {
		opOk = c.Database.InsertWatchEntry(
			ctx.PostForm("key"),
			ctx.PostForm("match-type"),
			ctx.PostForm("search-input"))
	} else if op == "delete" {
		opOk = c.Database.DeleteWatchEntry(ctx.PostForm("id")) == nil
	}

	ctx.HTML(http.StatusOK, "watches", gin.H{
		"path":    ctx.FullPath(),
		"op":      op,
		"opOk":    opOk,
		"results": c.Database.GetWatchEntries(),
	})
}
