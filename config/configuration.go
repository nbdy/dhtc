package config

import (
	"flag"
	"time"
)

type Configuration struct {
	DbName       string
	DatabaseType string
	DatabaseUrl  string
	Address      string
	MaxNeighbors uint          `form:"MaxNeighbors"`
	MaxLeeches   int           `form:"MaxLeeches"`
	DrainTimeout time.Duration `form:"DrainTimeout"`

	TelegramToken    string `form:"TelegramToken"`
	TelegramUsername string `form:"TelegramUsername"`

	DiscordWebhook string `form:"DiscordWebhook"`

	SlackWebhook string `form:"SlackWebhook"`

	GotifyURL   string `form:"GotifyURL"`
	GotifyToken string `form:"GotifyToken"`

	SafeMode bool `form:"SafeMode"`

	CrawlerThreads         int `form:"CrawlerThreads"`
	MaxConcurrentDownloads int `form:"MaxConcurrentDownloads"`
	RateLimit              int `form:"RateLimit"`

	EnableBlacklist bool   `form:"EnableBlacklist"`
	NameBlacklist   string `form:"NameBlacklist"`
	FileBlacklist   string `form:"FileBlacklist"`

	Statistics bool `form:"Statistics"`

	BootstrapNodeFile string `form:"BootstrapNodeFile"`

	OnlyWebServer bool
	AuthUser      string
	AuthPass      string
	SessionSecret string

	TransmissionURL  string `form:"TransmissionURL"`
	TransmissionUser string `form:"TransmissionUser"`
	TransmissionPass string `form:"TransmissionPass"`

	Aria2URL   string `form:"Aria2URL"`
	Aria2Token string `form:"Aria2Token"`

	DelugeURL  string `form:"DelugeURL"`
	DelugePass string `form:"DelugePass"`

	QBittorrentURL  string `form:"QBittorrentURL"`
	QBittorrentUser string `form:"QBittorrentUser"`
	QBittorrentPass string `form:"QBittorrentPass"`
}

func ParseArguments() *Configuration {
	config := Configuration{}

	flag.StringVar(&config.DbName, "database", "dhtdb", "database name (for CloverDB)")
	flag.StringVar(&config.DatabaseType, "database-type", "clover", "database type (clover, sqlite, postgres, mysql)")
	flag.StringVar(&config.DatabaseUrl, "database-url", "", "database URL (for GORM backends)")
	flag.StringVar(&config.Address, "address", ":4200", "address to run on")
	flag.UintVar(&config.MaxNeighbors, "MaxNeighbors", 500, "max. indexer neighbors")
	flag.IntVar(&config.MaxLeeches, "MaxLeeches", 128, "max. leeches")
	flag.DurationVar(&config.DrainTimeout, "DrainTimeout", 5*time.Second, "drain timeout")

	flag.StringVar(&config.TelegramToken, "TelegramToken", "", "bot token for notifications")
	flag.StringVar(&config.TelegramUsername, "TelegramUsername", "", "username to send notifications to")

	flag.StringVar(&config.DiscordWebhook, "DiscordWebhook", "", "Discord webhook URL")

	flag.StringVar(&config.SlackWebhook, "SlackWebhook", "", "Slack webhook URL")

	flag.StringVar(&config.GotifyURL, "GotifyURL", "", "Gotify URL")
	flag.StringVar(&config.GotifyToken, "GotifyToken", "", "Gotify token")

	flag.BoolVar(&config.SafeMode, "SafeMode", false, "start with safe mode enabled")
	flag.IntVar(&config.CrawlerThreads, "CrawlerThreads", 2, "dht crawler threads")
	flag.IntVar(&config.MaxConcurrentDownloads, "MaxConcurrentDownloads", 10, "max. concurrent metadata downloads")
	flag.IntVar(&config.RateLimit, "RateLimit", 100, "max. outgoing UDP packets per second per crawler")

	flag.BoolVar(&config.EnableBlacklist, "EnableBlacklist", false, "enable blacklists")
	flag.StringVar(&config.NameBlacklist, "NameBlacklist", "", "blacklist for torrent names")
	flag.StringVar(&config.FileBlacklist, "FileBlacklist", "", "blacklist for file names")

	flag.BoolVar(&config.Statistics, "Statistics", false, "enable Statistics (dashboard)")

	flag.StringVar(&config.BootstrapNodeFile, "BootstrapNodeFile", "bootstrap-nodes.txt", "bootstrap nodes to use")

	flag.BoolVar(&config.OnlyWebServer, "OnlyWebServer", false, "only start the web-server")
	flag.StringVar(&config.AuthUser, "auth-user", "", "username for basic auth")
	flag.StringVar(&config.AuthPass, "auth-pass", "", "password for basic auth")
	flag.StringVar(&config.SessionSecret, "session-secret", "dhtc-secret", "session secret")

	flag.StringVar(&config.TransmissionURL, "transmission-url", "", "Transmission RPC URL (e.g. http://localhost:9091/transmission/rpc)")
	flag.StringVar(&config.TransmissionUser, "transmission-user", "", "Transmission RPC username")
	flag.StringVar(&config.TransmissionPass, "transmission-pass", "", "Transmission RPC password")

	flag.StringVar(&config.Aria2URL, "aria2-url", "", "Aria2 RPC URL (e.g. http://localhost:6800/jsonrpc)")
	flag.StringVar(&config.Aria2Token, "aria2-token", "", "Aria2 RPC secret token")

	flag.StringVar(&config.DelugeURL, "deluge-url", "", "Deluge JSON-RPC URL (e.g. http://localhost:8112/json)")
	flag.StringVar(&config.DelugePass, "deluge-pass", "", "Deluge password")

	flag.StringVar(&config.QBittorrentURL, "qbittorrent-url", "", "qBittorrent URL (e.g. http://localhost:8080)")
	flag.StringVar(&config.QBittorrentUser, "qbittorrent-user", "", "qBittorrent username")
	flag.StringVar(&config.QBittorrentPass, "qbittorrent-pass", "", "qBittorrent password")

	flag.Parse()

	return &config
}
