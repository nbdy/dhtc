package ui

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) APISearch(ctx *gin.Context) {
	params := c.parseSearchParams(ctx)

	results, total, err := c.Database.Search(params.Key, params.MatchType, params.SearchInput, params.Limit, params.Offset, params.Filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total":       total,
		"currentPage": params.Page,
		"totalPages":  (total + int64(params.Limit) - 1) / int64(params.Limit),
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
	page, limit, offset := c.parsePagination(ctx)

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
