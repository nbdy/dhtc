package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (c *Controller) BlacklistGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "blacklist", gin.H{
		"path":    ctx.FullPath(),
		"results": c.Database.GetBlacklistEntries(),
	})
}

func (c *Controller) BlacklistPost(ctx *gin.Context) {
	opOk := false
	op := ctx.PostForm("op")
	if op == "add" {
		opOk = c.Database.AddToBlacklist([]string{ctx.PostForm("Filter")}, ctx.PostForm("Type"))
	} else if op == "delete" {
		opOk = c.Database.DeleteBlacklistItem(ctx.PostForm("Id")) == nil
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
		"results": c.Database.GetBlacklistEntries(),
	})
}
