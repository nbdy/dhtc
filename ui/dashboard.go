package ui

import (
	"dhtc/db"
	"github.com/gin-gonic/gin"
	"github.com/nikolalohinski/gonja"
	"net/http"
)

var dashboardTplBytes, _ = templates.ReadFile("ui/templates/dashboard.html")
var dashboardTpl = gonja.Must(gonja.FromBytes(dashboardTplBytes))

func (c *Controller) Dashboard(ctx *gin.Context) {
	out, _ := dashboardTpl.Execute(gonja.Context{
		"info_hash_count": db.GetInfoHashCount(c.Database),
		"path":            ctx.FullPath(),
		"statistics":      c.Configuration.Statistics,
	})
	ctx.Data(http.StatusOK, "text/html", []byte(out))
}
