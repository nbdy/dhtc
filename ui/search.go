package ui

import (
	"dhtc/db"
	"github.com/gin-gonic/gin"
	"github.com/nikolalohinski/gonja"
	"net/http"
)

var searchTplBytes, _ = templates.ReadFile("ui/templates/search.html")
var searchTpl = gonja.Must(gonja.FromBytes(searchTplBytes))

func (c *Controller) SearchGet(ctx *gin.Context) {
	out, _ := searchTpl.Execute(gonja.Context{"path": ctx.FullPath()})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}

func (c *Controller) SearchPost(ctx *gin.Context) {
	out, _ := searchTpl.Execute(gonja.Context{
		"results": db.FindBy(
			c.Database,
			ctx.PostForm("key"),
			ctx.PostForm("match-type"),
			ctx.PostForm("search-input")),
		"path": ctx.FullPath(),
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}
