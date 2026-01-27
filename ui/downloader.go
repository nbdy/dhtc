package ui

import (
	"dhtc/downloader"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SendToTransmission(ctx *gin.Context) {
	c.handleDownload(ctx, &downloader.TransmissionClient{
		URL:  c.Configuration.TransmissionURL,
		User: c.Configuration.TransmissionUser,
		Pass: c.Configuration.TransmissionPass,
	}, "transmission")
}

func (c *Controller) SendToAria2(ctx *gin.Context) {
	c.handleDownload(ctx, &downloader.Aria2Client{
		URL:   c.Configuration.Aria2URL,
		Token: c.Configuration.Aria2Token,
	}, "aria2")
}

func (c *Controller) SendToDeluge(ctx *gin.Context) {
	c.handleDownload(ctx, &downloader.DelugeClient{
		URL:  c.Configuration.DelugeURL,
		Pass: c.Configuration.DelugePass,
	}, "deluge")
}

func (c *Controller) SendToQBittorrent(ctx *gin.Context) {
	c.handleDownload(ctx, &downloader.QBittorrentClient{
		URL:  c.Configuration.QBittorrentURL,
		User: c.Configuration.QBittorrentUser,
		Pass: c.Configuration.QBittorrentPass,
	}, "qbittorrent")
}
