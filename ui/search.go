package ui

import (
	"dhtc/db"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SearchGet(ctx *gin.Context) {
	key := ctx.Query("key")
	matchType := ctx.Query("match-type")
	searchInput := ctx.Query("search-input")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	minSize, _ := strconv.ParseUint(ctx.DefaultQuery("min-size", "0"), 10, 64)
	maxSize, _ := strconv.ParseUint(ctx.DefaultQuery("max-size", "0"), 10, 64)
	startDateVal := ctx.Query("start-date-val")
	endDateVal := ctx.Query("end-date-val")

	var startDate, endDate int64
	if startDateVal != "" {
		t, _ := time.Parse("2006-01-02", startDateVal)
		startDate = t.Unix()
	}
	if endDateVal != "" {
		t, _ := time.Parse("2006-01-02", endDateVal)
		endDate = t.Unix()
	}

	filters := db.SearchFilters{
		MinSize:   minSize,
		MaxSize:   maxSize,
		StartDate: startDate,
		EndDate:   endDate,
	}

	if searchInput == "" {
		ctx.HTML(http.StatusOK, "search", c.getCommonH(ctx))
		return
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	results, total, _ := c.Database.Search(key, matchType, searchInput, limit, offset, filters)

	h := c.getCommonH(ctx)
	h["results"] = results
	h["currentPage"] = page
	h["totalPages"] = (total + int64(limit) - 1) / int64(limit)
	h["limit"] = limit
	h["total"] = total
	h["key"] = key
	h["matchType"] = matchType
	h["searchInput"] = searchInput
	h["minSize"] = minSize
	h["maxSize"] = maxSize
	h["startDateVal"] = startDateVal
	h["endDateVal"] = endDateVal

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
