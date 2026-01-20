package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SettingsGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "settings", c.getCommonH(ctx))
}

func (c *Controller) SettingsPost(ctx *gin.Context) {
	if err := ctx.ShouldBind(c.Configuration); err != nil {
		h := c.getCommonH(ctx)
		h["error"] = err.Error()
		ctx.HTML(http.StatusBadRequest, "settings", h)
		return
	}
	h := c.getCommonH(ctx)
	h["saved"] = true
	ctx.HTML(http.StatusOK, "settings", h)
}
