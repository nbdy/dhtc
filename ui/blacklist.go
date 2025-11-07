package ui

import (
	"dhtc/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (c *Controller) BlacklistGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "blacklist", gin.H{
		"path":    ctx.FullPath(),
		"results": db.GetBlacklistEntries(c.Database),
	})
}

func (c *Controller) BlacklistPost(ctx *gin.Context) {
	opOk := false
	op := ctx.PostForm("op")
	if op == "add" {
		opOk = db.AddToBlacklist(c.Database, []string{ctx.PostForm("Filter")}, ctx.PostForm("Type"))
	} else if op == "delete" {
		opOk = db.DeleteBlacklistItem(c.Database, ctx.PostForm("Id"))
	} else if op == "enable" {
		c.Configuration.EnableBlacklist = true
		opOk = true
	} else if op == "disable" {
		c.Configuration.EnableBlacklist = false
		opOk = true
	}

	log.Print(opOk)

	ctx.HTML(http.StatusOK, "blacklist", gin.H{
		"path":    ctx.FullPath(),
		"op":      op,
		"opOk":    opOk,
		"results": db.GetBlacklistEntries(c.Database),
	})
}
