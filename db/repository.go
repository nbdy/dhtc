package db

import (
	dhtcclient "dhtc/dhtc-client"
)

type Repository interface {
	GetInfoHashCount() int
	FindBy(key string, searchType string, searchInput string) []MetaData
	Search(key string, searchType string, searchInput string, limit int, offset int, filters SearchFilters) ([]MetaData, int64, error)
	GetNRandomEntries(n int) []MetaData
	GetLatest(limit int, offset int) ([]MetaData, int64, error)
	InsertMetadata(md dhtcclient.Metadata) bool

	GetWatchEntries() []WatchEntry
	InsertWatchEntry(key string, searchType string, searchInput string) bool
	DeleteWatchEntry(id string) error

	GetBlacklistEntries() []BlacklistEntry
	AddToBlacklist(filters []string, entryType string) bool
	DeleteBlacklistItem(id string) error
	IsBlacklisted(md dhtcclient.Metadata) bool

	GetAllInfoHashes() ([]string, error)

	GetStats(limit int) ([]Stats, error)
	GetStatsByInterval(interval string, limit int) ([]Stats, error)
	InsertStats(stats Stats) error

	GetCategoryDistribution() (map[string]int64, error)

	Close() error
}
