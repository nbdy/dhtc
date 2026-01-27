package ui

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SearchGet(ctx *gin.Context) {
	params := c.parseSearchParams(ctx)

	if params.SearchInput == "" {
		ctx.HTML(http.StatusOK, "search", c.getCommonH(ctx))
		return
	}

	results, total, _ := c.Database.Search(params.Key, params.MatchType, params.SearchInput, params.Limit, params.Offset, params.Filters)

	h := c.getCommonH(ctx)
	h["results"] = results
	h["currentPage"] = params.Page
	h["totalPages"] = (total + int64(params.Limit) - 1) / int64(params.Limit)
	h["limit"] = params.Limit
	h["total"] = total
	h["key"] = params.Key
	h["matchType"] = params.MatchType
	h["searchInput"] = params.SearchInput
	h["minSize"] = params.Filters.MinSize
	h["maxSize"] = params.Filters.MaxSize
	h["startDateVal"] = params.StartDateVal
	h["endDateVal"] = params.EndDateVal

	ctx.HTML(http.StatusOK, "search", h)
}

func (c *Controller) SearchPost(ctx *gin.Context) {
	key := ctx.PostForm("key")
	matchType := ctx.PostForm("match-type")
	searchInput := ctx.PostForm("search-input")
	minSize := ctx.PostForm("min-size")
	maxSize := ctx.PostForm("max-size")
	startDateVal := ctx.PostForm("start-date-val")
	endDateVal := ctx.PostForm("end-date-val")

	params := url.Values{}
	params.Add("key", key)
	params.Add("match-type", matchType)
	params.Add("search-input", searchInput)
	if minSize != "0" && minSize != "" {
		params.Add("min-size", minSize)
	}
	if maxSize != "0" && maxSize != "" {
		params.Add("max-size", maxSize)
	}
	if startDateVal != "" {
		params.Add("start-date-val", startDateVal)
	}
	if endDateVal != "" {
		params.Add("end-date-val", endDateVal)
	}

	ctx.Redirect(http.StatusSeeOther, "/search?"+params.Encode())
}
