package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SettingsGet(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "settings", c.getCommonH(ctx))
}
