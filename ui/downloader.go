package ui

import (
	"dhtc/downloader"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SendToTransmission(ctx *gin.Context) {
	magnet := ctx.Query("magnet")
	if magnet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "magnet is required"})
		return
	}

	client := downloader.TransmissionClient{
		URL:  c.Configuration.TransmissionURL,
		User: c.Configuration.TransmissionUser,
		Pass: c.Configuration.TransmissionPass,
	}

	err := client.AddMagnet(magnet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "sent to transmission"})
}

func (c *Controller) SendToAria2(ctx *gin.Context) {
	magnet := ctx.Query("magnet")
	if magnet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "magnet is required"})
		return
	}

	client := downloader.Aria2Client{
		URL:   c.Configuration.Aria2URL,
		Token: c.Configuration.Aria2Token,
	}

	err := client.AddMagnet(magnet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "sent to aria2"})
}

func (c *Controller) SendToDeluge(ctx *gin.Context) {
	magnet := ctx.Query("magnet")
	if magnet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "magnet is required"})
		return
	}

	client := downloader.DelugeClient{
		URL:  c.Configuration.DelugeURL,
		Pass: c.Configuration.DelugePass,
	}

	err := client.AddMagnet(magnet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "sent to deluge"})
}

func (c *Controller) SendToQBittorrent(ctx *gin.Context) {
	magnet := ctx.Query("magnet")
	if magnet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "magnet is required"})
		return
	}

	client := downloader.QBittorrentClient{
		URL:  c.Configuration.QBittorrentURL,
		User: c.Configuration.QBittorrentUser,
		Pass: c.Configuration.QBittorrentPass,
	}

	err := client.AddMagnet(magnet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "sent to qbittorrent"})
}
