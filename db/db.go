package db

import (
	"dhtc/config"
	"time"
)

const WatchTable = "watches"
const BlacklistTable = "blacklist"
const TorrentTable = "torrents"
const StatsTable = "stats"

func OpenRepository(cfg *config.Configuration) (Repository, error) {
	if cfg.DatabaseType == "clover" {
		return NewCloverRepository(cfg)
	}
	return NewGormRepository(cfg, cfg.DatabaseType, cfg.DatabaseUrl)
}

func GetIntervalDuration(interval string) time.Duration {
	switch interval {
	case "minute":
		return time.Minute
	case "hour":
		return time.Hour
	case "day":
		return 24 * time.Hour
	default:
		return time.Hour
	}
}
