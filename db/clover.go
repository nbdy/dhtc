package db

import (
	"dhtc/config"
	dhtcclient "dhtc/dhtc-client"
	"encoding/hex"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"regexp"
	"time"
)

type CloverRepository struct {
	db     *clover.DB
	config *config.Configuration
}

func NewCloverRepository(config *config.Configuration) (Repository, error) {
	err := os.MkdirAll(config.DbName, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msgf("Could not create database directory: %s", config.DbName)
	}

	db, err := clover.Open(config.DbName)
	if err != nil {
		return nil, err
	}

	_ = db.CreateCollection(TorrentTable)
	_ = db.CreateCollection(WatchTable)
	_ = db.CreateCollection(BlacklistTable)
	_ = db.CreateCollection(StatsTable)

	return &CloverRepository{
		db:     db,
		config: config,
	}, nil
}

func (r *CloverRepository) GetInfoHashCount() int {
	vals, _ := r.db.FindAll(query.NewQuery(TorrentTable))
	return len(vals)
}

func (r *CloverRepository) FindBy(key string, searchType string, searchInput string) []MetaData {
	values, _ := r.db.FindAll(query.NewQuery(TorrentTable).MatchFunc(func(doc *document.Document) bool {
		return Matches(doc, key, searchType, searchInput)
	}))
	return Documents2MetaData(values)
}

func (r *CloverRepository) Search(key string, searchType string, searchInput string, limit int, offset int, filters SearchFilters) ([]MetaData, int64, error) {
	q := query.NewQuery(TorrentTable).MatchFunc(func(doc *document.Document) bool {
		if !Matches(doc, key, searchType, searchInput) {
			return false
		}
		if filters.MinSize > 0 && doc.Get("TotalSize").(uint64) < filters.MinSize {
			return false
		}
		if filters.MaxSize > 0 && doc.Get("TotalSize").(uint64) > filters.MaxSize {
			return false
		}
		if filters.StartDate > 0 && doc.Get("DiscoveredOn").(int64) < filters.StartDate {
			return false
		}
		if filters.EndDate > 0 && doc.Get("DiscoveredOn").(int64) > filters.EndDate {
			return false
		}
		return true
	})

	total, _ := r.db.Count(q)
	values, err := r.db.FindAll(q.Sort(query.SortOption{Field: "DiscoveredOn", Direction: -1}).Limit(limit).Skip(offset))
	if err != nil {
		return nil, 0, err
	}
	return Documents2MetaData(values), int64(total), nil
}

func (r *CloverRepository) GetNRandomEntries(N int) []MetaData {
	count, _ := r.db.Count(query.NewQuery(TorrentTable))
	if count < N {
		N = count
	}

	all, _ := r.db.FindAll(query.NewQuery(TorrentTable))
	rVal := make([]*document.Document, N)
	for i := 0; i < N; i++ {
		rVal[i] = all[rand.Intn(count)]
	}
	return Documents2MetaData(rVal)
}

func (r *CloverRepository) GetLatest(limit int, offset int) ([]MetaData, int64, error) {
	q := query.NewQuery(TorrentTable)
	total, _ := r.db.Count(q)
	values, err := r.db.FindAll(q.Sort(query.SortOption{Field: "DiscoveredOn", Direction: -1}).Limit(limit).Skip(offset))
	if err != nil {
		return nil, 0, err
	}
	return Documents2MetaData(values), int64(total), nil
}

func (r *CloverRepository) InsertMetadata(md dhtcclient.Metadata) bool {
	if r.config.EnableBlacklist && r.IsBlacklisted(md) {
		log.Info().Msgf("Blacklisted: %s", md.Name)
		return false
	}

	doc := document.NewDocument()
	doc.Set("Name", md.Name)
	doc.Set("InfoHash", hex.EncodeToString(md.InfoHash))
	doc.Set("Files", md.Files)
	doc.Set("DiscoveredOn", md.DiscoveredOn)
	doc.Set("TotalSize", md.TotalSize)
	doc.Set("Categories", Categorize(md))
	_, err := r.db.InsertOne(TorrentTable, doc)
	return err == nil
}

func (r *CloverRepository) GetWatchEntries() []WatchEntry {
	all, _ := r.db.FindAll(query.NewQuery(WatchTable))
	rVal := make([]WatchEntry, len(all))
	for i, value := range all {
		rVal[i] = WatchEntry{
			value.ObjectId(),
			value.Get("Key").(string),
			value.Get("MatchType").(string),
			value.Get("Content").(string),
		}
	}
	return rVal
}

func (r *CloverRepository) InsertWatchEntry(key string, searchType string, searchInput string) bool {
	doc := document.NewDocument()
	doc.Set("Key", key)
	doc.Set("MatchType", searchType)
	doc.Set("Content", searchInput)
	_, err := r.db.InsertOne(WatchTable, doc)
	if err != nil {
		log.Error().Err(err).Msg("Could not insert watch entry")
		return false
	}
	return true
}

func (r *CloverRepository) DeleteWatchEntry(entryId string) error {
	return r.db.DeleteById(WatchTable, entryId)
}

func (r *CloverRepository) GetBlacklistEntries() []BlacklistEntry {
	all, _ := r.db.FindAll(query.NewQuery(BlacklistTable))
	rVal := make([]BlacklistEntry, len(all))
	for i, value := range all {
		rVal[i] = BlacklistEntry{
			value.ObjectId(),
			value.Get("Filter").(string),
			GetBlacklistTypeFromStrInt(value.Get("Type").(string)),
		}
	}
	return rVal
}

func (r *CloverRepository) AddToBlacklist(filters []string, entryType string) bool {
	rVal := false
	for _, filter := range filters {
		if !r.isInBlacklist(filter, entryType) {
			doc := document.NewDocument()
			doc.Set("Type", entryType)
			doc.Set("Filter", filter)
			_, err := r.db.InsertOne(BlacklistTable, doc)
			if err != nil {
				log.Error().Err(err).Msg("Could not add to blacklist")
			}
			rVal = err == nil
		}
	}
	return rVal
}

func (r *CloverRepository) isInBlacklist(filter string, entryType string) bool {
	result, _ := r.db.FindAll(query.NewQuery(BlacklistTable).Where(query.Field("Filter").Eq(filter).And(query.Field("Type").Eq(entryType))))
	return len(result) > 0
}

func (r *CloverRepository) DeleteBlacklistItem(itemId string) error {
	return r.db.DeleteById(BlacklistTable, itemId)
}

func (r *CloverRepository) IsBlacklisted(md dhtcclient.Metadata) bool {
	all, _ := r.db.FindAll(query.NewQuery(BlacklistTable).MatchFunc(func(doc *document.Document) bool {
		filterStr := doc.Get("Filter").(string)
		filter, err := regexp.Compile(filterStr)
		if err != nil {
			return false
		}

		switch doc.Get("Type").(string) {
		case "0":
			return filter.MatchString(md.Name)
		case "1":
			return IsFileBlacklisted(md, filter)
		}

		return false
	}))
	return len(all) > 0
}

func (r *CloverRepository) GetAllInfoHashes() ([]string, error) {
	all, err := r.db.FindAll(query.NewQuery(TorrentTable))
	if err != nil {
		return nil, err
	}
	res := make([]string, len(all))
	for i, d := range all {
		res[i] = d.Get("InfoHash").(string)
	}
	return res, nil
}

func (r *CloverRepository) GetStats(limit int) ([]Stats, error) {
	q := query.NewQuery(StatsTable).Sort(query.SortOption{Field: "Timestamp", Direction: -1}).Limit(limit)
	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}
	rVal := make([]Stats, len(docs))
	for i, doc := range docs {
		rVal[i] = Stats{
			Timestamp:    doc.Get("Timestamp").(time.Time),
			TorrentCount: doc.Get("TorrentCount").(int64),
		}
	}
	return rVal, nil
}

func (r *CloverRepository) GetStatsByInterval(interval string, limit int) ([]Stats, error) {
	duration := GetIntervalDuration(interval)
	now := time.Now().Truncate(duration)
	since := now.Add(-time.Duration(limit) * duration).Unix()

	q := query.NewQuery(TorrentTable).Where(query.Field("DiscoveredOn").GtEq(since))
	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}

	bucketMap := make(map[int64]int64)
	seconds := int64(duration.Seconds())
	for _, doc := range docs {
		discoveredOn := doc.Get("DiscoveredOn").(int64)
		bucket := (discoveredOn / seconds) * seconds
		bucketMap[bucket]++
	}

	res := make([]Stats, limit)
	for i := 0; i < limit; i++ {
		ts := now.Add(-time.Duration(i) * duration)
		res[i] = Stats{
			Timestamp:    ts,
			TorrentCount: bucketMap[ts.Unix()],
		}
	}
	return res, nil
}

func (r *CloverRepository) InsertStats(stats Stats) error {
	doc := document.NewDocument()
	doc.Set("Timestamp", stats.Timestamp)
	doc.Set("TorrentCount", stats.TorrentCount)
	_, err := r.db.InsertOne(StatsTable, doc)
	return err
}

func (r *CloverRepository) GetCategoryDistribution() (map[string]int64, error) {
	docs, err := r.db.FindAll(query.NewQuery(TorrentTable))
	if err != nil {
		return nil, err
	}

	dist := make(map[string]int64)
	for _, doc := range docs {
		var categories []string
		if doc.Has("Categories") {
			switch v := doc.Get("Categories").(type) {
			case []string:
				categories = v
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok {
						categories = append(categories, s)
					}
				}
			}
		} else if doc.Has("Category") {
			categories = []string{doc.Get("Category").(string)}
		}

		if len(categories) == 0 {
			dist["unknown"]++
		} else {
			for _, cat := range categories {
				dist[cat]++
			}
		}
	}
	return dist, nil
}

func (r *CloverRepository) Close() error {
	return r.db.Close()
}
