package db

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

type MetaData struct {
	Name         string
	InfoHash     string
	DiscoveredOn string
	TotalSize    uint64
	Files        []interface{}
}
