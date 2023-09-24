package ui

import (
	"dhtc/db"
	"github.com/gin-gonic/gin"
	"github.com/nikolalohinski/gonja"
	"net/http"
	"strconv"
)

var discoverTplBytes, _ = templates.ReadFile("ui/templates/discover.html")
var discoverTpl = gonja.Must(gonja.FromBytes(discoverTplBytes))

func (c *Controller) DiscoverGet(ctx *gin.Context) {
	out, _ := discoverTpl.Execute(gonja.Context{
		"results": db.GetNRandomEntries(c.Database, 50),
		"path":    ctx.FullPath(),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}

func (c *Controller) DiscoverPost(ctx *gin.Context) {
	N, err := strconv.Atoi(ctx.PostForm("limit"))
	if err != nil {
		N = 50
	}
	out, _ := discoverTpl.Execute(gonja.Context{
		"results": db.GetNRandomEntries(c.Database, N),
		"path":    ctx.FullPath(),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}
