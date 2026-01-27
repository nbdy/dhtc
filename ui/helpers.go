package ui

import (
	"dhtc/db"
	"dhtc/downloader"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchParams struct {
	Key          string
	MatchType    string
	SearchInput  string
	Page         int
	Limit        int
	Offset       int
	StartDateVal string
	EndDateVal   string
	Filters      db.SearchFilters
}

func (c *Controller) parsePagination(ctx *gin.Context) (page, limit, offset int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset = (page - 1) * limit
	return
}

func (c *Controller) parseSearchParams(ctx *gin.Context) SearchParams {
	key := ctx.Query("key")
	matchType := ctx.Query("match-type")
	searchInput := ctx.Query("search-input")
	page, limit, offset := c.parsePagination(ctx)

	minSize, _ := strconv.ParseUint(ctx.DefaultQuery("min-size", "0"), 10, 64)
	maxSize, _ := strconv.ParseUint(ctx.DefaultQuery("max-size", "0"), 10, 64)

	startDateVal := ctx.Query("start-date-val")
	endDateVal := ctx.Query("end-date-val")

	var startDate, endDate int64
	if startDateVal != "" {
		t, _ := time.Parse("2006-01-02", startDateVal)
		startDate = t.Unix()
	} else {
		startDate, _ = strconv.ParseInt(ctx.DefaultQuery("start-date", "0"), 10, 64)
	}

	if endDateVal != "" {
		t, _ := time.Parse("2006-01-02", endDateVal)
		endDate = t.Unix()
	} else {
		endDate, _ = strconv.ParseInt(ctx.DefaultQuery("end-date", "0"), 10, 64)
	}

	return SearchParams{
		Key:          key,
		MatchType:    matchType,
		SearchInput:  searchInput,
		Page:         page,
		Limit:        limit,
		Offset:       offset,
		StartDateVal: startDateVal,
		EndDateVal:   endDateVal,
		Filters: db.SearchFilters{
			MinSize:   minSize,
			MaxSize:   maxSize,
			StartDate: startDate,
			EndDate:   endDate,
		},
	}
}

func (c *Controller) handleDownload(ctx *gin.Context, client downloader.Client, name string) {
	magnet := ctx.Query("magnet")
	if magnet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "magnet is required"})
		return
	}

	err := client.AddMagnet(magnet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "sent to " + name})
}
