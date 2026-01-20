package ui

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) DiscoverGet(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	results, total, _ := c.Database.GetLatest(limit, offset)

	h := c.getCommonH(ctx)
	h["results"] = results
	h["currentPage"] = page
	h["totalPages"] = (total + int64(limit) - 1) / int64(limit)
	h["limit"] = limit
	h["total"] = total

	ctx.HTML(http.StatusOK, "discover", h)
}

func (c *Controller) DiscoverPost(ctx *gin.Context) {
	N, err := strconv.Atoi(ctx.PostForm("limit"))
	if err != nil {
		N = 50
	}

	h := c.getCommonH(ctx)
	h["results"] = c.Database.GetNRandomEntries(N)

	ctx.HTML(http.StatusOK, "discover", h)
}
