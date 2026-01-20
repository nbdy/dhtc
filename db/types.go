package db

import "time"

type Stats struct {
	Timestamp    time.Time `gorm:"primaryKey"`
	TorrentCount int64
}

type WatchEntry struct {
	Id        string
	Key       string
	MatchType string
	Content   string
}

type BlacklistEntry struct {
	Id     string
	Filter string
	Type   string
}

type SearchFilters struct {
	MinSize   uint64
	MaxSize   uint64
	StartDate int64
	EndDate   int64
}

type MetaData struct {
	Name         string
	InfoHash     string
	DiscoveredOn string
	TotalSize    uint64
	Files        []interface{}
	Categories   []string
}
