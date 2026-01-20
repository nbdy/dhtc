package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Trawl(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "trawl", c.getCommonH(ctx))
}
