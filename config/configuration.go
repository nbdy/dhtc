package config

import (
	"flag"
	"time"
)

type Configuration struct {
	DbName       string
	Address      string
	MaxNeighbors uint
	MaxLeeches   int
	DrainTimeout time.Duration

	TelegramToken    string
	TelegramUsername string

	SafeMode bool

	CrawlerThreads int

	EnableBlacklist bool
	NameBlacklist   string
	FileBlacklist   string

	Statistics bool

	BootstrapNodeFile string

	OnlyWebServer bool
}

func ParseArguments() *Configuration {
	config := Configuration{}

	flag.StringVar(&config.DbName, "database", "dhtdb", "database name")
	flag.StringVar(&config.Address, "address", ":4200", "address to run on")
	flag.UintVar(&config.MaxNeighbors, "MaxNeighbors", 1000, "max. indexer neighbors")
	flag.IntVar(&config.MaxLeeches, "MaxLeeches", 256, "max. leeches")
	flag.DurationVar(&config.DrainTimeout, "DrainTimeout", 5*time.Second, "drain timeout")

	flag.StringVar(&config.TelegramToken, "TelegramToken", "", "bot token for notifications")
	flag.StringVar(&config.TelegramUsername, "TelegramUsername", "", "username to send notifications to")

	flag.BoolVar(&config.SafeMode, "SafeMode", false, "start with safe mode enabled")
	flag.IntVar(&config.CrawlerThreads, "CrawlerThreads", 5, "dht crawler threads")

	flag.BoolVar(&config.EnableBlacklist, "EnableBlacklist", false, "enable blacklists")
	flag.StringVar(&config.NameBlacklist, "NameBlacklist", "", "blacklist for torrent names")
	flag.StringVar(&config.FileBlacklist, "FileBlacklist", "", "blacklist for file names")

	flag.BoolVar(&config.Statistics, "Statistics", false, "enable Statistics (dashboard)")

	flag.StringVar(&config.BootstrapNodeFile, "BootstrapNodeFile", "bootstrap-nodes.txt", "bootstrap nodes to use")

	flag.BoolVar(&config.OnlyWebServer, "OnlyWebServer", false, "only start the web-server")

	flag.Parse()

	return &config
}
