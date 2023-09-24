package ui

import (
	"dhtc/db"
	"github.com/gin-gonic/gin"
	"github.com/nikolalohinski/gonja"
	"net/http"
)

var blacklistTplBytes, _ = templates.ReadFile("ui/templates/blacklist.html")
var blacklistTpl = gonja.Must(gonja.FromBytes(blacklistTplBytes))

func (c *Controller) BlacklistGet(ctx *gin.Context) {
	out, _ := blacklistTpl.Execute(gonja.Context{
		"path":    ctx.FullPath(),
		"results": db.GetBlacklistEntries(c.Database),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
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
	out, _ := blacklistTpl.Execute(gonja.Context{
		"path":    ctx.FullPath(),
		"op":      op,
		"opOk":    opOk,
		"results": db.GetBlacklistEntries(c.Database),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}
