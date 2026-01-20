package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Dashboard(ctx *gin.Context) {
	catDist, _ := c.Database.GetCategoryDistribution()
	h := c.getCommonH(ctx)
	h["info_hash_count"] = c.Database.GetInfoHashCount()
	h["statistics"] = c.Configuration.Statistics
	h["catDist"] = catDist
	ctx.HTML(http.StatusOK, "dashboard", h)
}
