package ui

import (
	"dhtc/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) WatchGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "watches", gin.H{
		"path":    ctx.FullPath(),
		"results": db.GetWatchEntries(c.Database),
	})
}

func (c *Controller) WatchPost(ctx *gin.Context) {
	opOk := false
	op := ctx.PostForm("op")
	if op == "add" {
		opOk = db.InsertWatchEntry(
			c.Database,
			ctx.PostForm("key"),
			ctx.PostForm("match-type"),
			ctx.PostForm("search-input"))
	} else if op == "delete" {
		opOk = db.DeleteWatchEntry(c.Database, ctx.PostForm("id"))
	}

	ctx.HTML(http.StatusOK, "watches", gin.H{
		"path":    ctx.FullPath(),
		"op":      op,
		"opOk":    opOk,
		"results": db.GetWatchEntries(c.Database),
	})
}
