package ui

import (
	"dhtc/db"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) APISearch(ctx *gin.Context) {
	key := ctx.Query("key")
	matchType := ctx.Query("match-type")
	searchInput := ctx.Query("search-input")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	minSize, _ := strconv.ParseUint(ctx.DefaultQuery("min-size", "0"), 10, 64)
	maxSize, _ := strconv.ParseUint(ctx.DefaultQuery("max-size", "0"), 10, 64)
	startDate, _ := strconv.ParseInt(ctx.DefaultQuery("start-date", "0"), 10, 64)
	endDate, _ := strconv.ParseInt(ctx.DefaultQuery("end-date", "0"), 10, 64)

	filters := db.SearchFilters{
		MinSize:   minSize,
		MaxSize:   maxSize,
		StartDate: startDate,
		EndDate:   endDate,
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	results, total, err := c.Database.Search(key, matchType, searchInput, limit, offset, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total":       total,
		"currentPage": page,
		"totalPages":  (total + int64(limit) - 1) / int64(limit),
	})
}

func (c *Controller) APIStats(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "24"))
	interval := ctx.DefaultQuery("interval", "hour")
	stats, err := c.Database.GetStatsByInterval(interval, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"stats":           stats,
		"interval":        interval,
		"info_hash_count": c.Database.GetInfoHashCount(),
	})
}

func (c *Controller) APICategories(ctx *gin.Context) {
	dist, err := c.Database.GetCategoryDistribution()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dist)
}

func (c *Controller) APILatest(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	results, total, err := c.Database.GetLatest(limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total":       total,
		"currentPage": page,
		"totalPages":  (total + int64(limit) - 1) / int64(limit),
	})
}
