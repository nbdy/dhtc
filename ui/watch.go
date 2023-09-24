package ui

import (
	"dhtc/db"
	"github.com/gin-gonic/gin"
	"github.com/nikolalohinski/gonja"
	"net/http"
)

var watchTplBytes, _ = templates.ReadFile("ui/templates/watches.html")
var watchTpl = gonja.Must(gonja.FromBytes(watchTplBytes))

func (c *Controller) WatchGet(ctx *gin.Context) {
	out, _ := watchTpl.Execute(gonja.Context{
		"path":    ctx.FullPath(),
		"results": db.GetWatchEntries(c.Database),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
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
	out, _ := watchTpl.Execute(gonja.Context{
		"path":    ctx.FullPath(),
		"op":      op,
		"opOk":    opOk,
		"results": db.GetWatchEntries(c.Database),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}
