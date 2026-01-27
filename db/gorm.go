package db

import (
	"dhtc/config"
	dhtcclient "dhtc/dhtc-client"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormTorrent struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"index"`
	InfoHash     string `gorm:"uniqueIndex"`
	Files        string `gorm:"type:text"`
	DiscoveredOn int64  `gorm:"index"`
	TotalSize    uint64
	Categories   string `gorm:"column:category;index"`
}

type GormWatch struct {
	ID        uint `gorm:"primaryKey"`
	Key       string
	MatchType string
	Content   string
}

type GormBlacklist struct {
	ID     uint `gorm:"primaryKey"`
	Filter string
	Type   string
}

type GormStats struct {
	Timestamp    time.Time `gorm:"primaryKey"`
	TorrentCount int64
}

type GormRepository struct {
	db         *gorm.DB
	config     *config.Configuration
	ftsEnabled bool
}

func NewGormRepository(config *config.Configuration, dbType, dbUrl string) (Repository, error) {
	var dialector gorm.Dialector

	switch strings.ToLower(dbType) {
	case "sqlite":
		dialector = sqlite.Open(dbUrl)
	case "postgres", "postgresql":
		dialector = postgres.Open(dbUrl)
	case "mysql":
		dialector = mysql.Open(dbUrl)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&GormTorrent{}, &GormWatch{}, &GormBlacklist{}, &GormStats{})
	if err != nil {
		return nil, err
	}

	repo := &GormRepository{
		db:     db,
		config: config,
	}

	repo.initFTS()

	return repo, nil
}

func (r *GormRepository) initFTS() {
	dialector := r.db.Dialector.Name()
	switch dialector {
	case "mysql":
		var count int64
		r.db.Raw("SELECT COUNT(1) FROM information_schema.statistics WHERE table_name = 'gorm_torrents' AND index_name = 'idx_name_fts'").Scan(&count)
		if count == 0 {
			if err := r.db.Exec("ALTER TABLE gorm_torrents ADD FULLTEXT INDEX idx_name_fts (name)").Error; err != nil {
				log.Error().Err(err).Msg("failed to create fulltext index for mysql")
				return
			}
		}
		r.ftsEnabled = true
	case "postgres":
		if err := r.db.Exec("CREATE INDEX IF NOT EXISTS idx_name_fts ON gorm_torrents USING gin(to_tsvector('simple', name))").Error; err != nil {
			log.Error().Err(err).Msg("failed to create gin index for postgres")
			return
		}
		r.ftsEnabled = true
	case "sqlite":
		if err := r.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS gorm_torrents_fts USING fts5(name, content='gorm_torrents', content_rowid='id')").Error; err != nil {
			log.Warn().Err(err).Msg("failed to create fts5 virtual table for sqlite, FTS will be disabled")
			return
		}
		// Triggers to keep FTS in sync
		r.db.Exec(`CREATE TRIGGER IF NOT EXISTS gorm_torrents_ai AFTER INSERT ON gorm_torrents BEGIN
		  INSERT INTO gorm_torrents_fts(rowid, name) VALUES (new.id, new.name);
		END;`)
		r.db.Exec(`CREATE TRIGGER IF NOT EXISTS gorm_torrents_ad AFTER DELETE ON gorm_torrents BEGIN
		  INSERT INTO gorm_torrents_fts(gorm_torrents_fts, rowid, name) VALUES('delete', old.id, old.name);
		END;`)
		r.db.Exec(`CREATE TRIGGER IF NOT EXISTS gorm_torrents_au AFTER UPDATE ON gorm_torrents BEGIN
		  INSERT INTO gorm_torrents_fts(gorm_torrents_fts, rowid, name) VALUES('delete', old.id, old.name);
		  INSERT INTO gorm_torrents_fts(rowid, name) VALUES (new.id, new.name);
		END;`)

		var count int64
		r.db.Raw("SELECT count(*) FROM gorm_torrents_fts").Scan(&count)
		if count == 0 {
			r.db.Exec("INSERT INTO gorm_torrents_fts(rowid, name) SELECT id, name FROM gorm_torrents")
		}
		r.ftsEnabled = true
	}
}

func (r *GormRepository) GetInfoHashCount() int {
	var count int64
	r.db.Model(&GormTorrent{}).Count(&count)
	return int(count)
}

func (r *GormRepository) applyNameSearch(query *gorm.DB, searchType string, searchInput string) *gorm.DB {
	switch searchType {
	case "contains":
		if r.ftsEnabled {
			dialector := r.db.Dialector.Name()
			switch dialector {
			case "mysql":
				return query.Where("MATCH(name) AGAINST(? IN BOOLEAN MODE)", searchInput+"*")
			case "postgres":
				return query.Where("to_tsvector('simple', name) @@ plainto_tsquery('simple', ?)", searchInput)
			case "sqlite":
				return query.Where("id IN (SELECT rowid FROM gorm_torrents_fts WHERE name MATCH ?)", searchInput+"*")
			}
		}
		return query.Where("name LIKE ?", "%"+searchInput+"%")
	case "equals":
		return query.Where("name = ?", searchInput)
	case "startswith":
		return query.Where("name LIKE ?", searchInput+"%")
	case "endswith":
		return query.Where("name LIKE ?", "%"+searchInput)
	}
	return query
}

func (r *GormRepository) FindBy(key string, searchType string, searchInput string) []MetaData {
	var torrents []GormTorrent
	query := r.db.Model(&GormTorrent{})

	switch key {
	case "Name":
		query = r.applyNameSearch(query, searchType, searchInput)
	case "InfoHash":
		query = query.Where("info_hash = ?", searchInput)
	case "Files":
		query = query.Where("files LIKE ?", "%"+searchInput+"%")
	}

	query.Find(&torrents)

	return r.toMetaDataSlice(torrents)
}

func (r *GormRepository) Search(key string, searchType string, searchInput string, limit int, offset int, filters SearchFilters) ([]MetaData, int64, error) {
	var torrents []GormTorrent
	query := r.db.Model(&GormTorrent{})

	switch key {
	case "Name":
		query = r.applyNameSearch(query, searchType, searchInput)
	case "InfoHash":
		query = query.Where("info_hash = ?", searchInput)
	case "Files":
		query = query.Where("files LIKE ?", "%"+searchInput+"%")
	case "DiscoveredOn":
		query = query.Where("discovered_on LIKE ?", "%"+searchInput+"%")
	}

	if filters.MinSize > 0 {
		query = query.Where("total_size >= ?", filters.MinSize)
	}
	if filters.MaxSize > 0 {
		query = query.Where("total_size <= ?", filters.MaxSize)
	}
	if filters.StartDate > 0 {
		query = query.Where("discovered_on >= ?", filters.StartDate)
	}
	if filters.EndDate > 0 {
		query = query.Where("discovered_on <= ?", filters.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("discovered_on DESC").Limit(limit).Offset(offset).Find(&torrents).Error; err != nil {
		return nil, 0, err
	}

	return r.toMetaDataSlice(torrents), total, nil
}

func (r *GormRepository) GetNRandomEntries(n int) []MetaData {
	var torrents []GormTorrent
	order := "RANDOM()"
	if r.db.Dialector.Name() == "mysql" {
		order = "RAND()"
	}
	r.db.Order(order).Limit(n).Find(&torrents)
	return r.toMetaDataSlice(torrents)
}

func (r *GormRepository) GetLatest(limit int, offset int) ([]MetaData, int64, error) {
	var torrents []GormTorrent
	var total int64

	if err := r.db.Model(&GormTorrent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Model(&GormTorrent{}).Order("discovered_on DESC").Limit(limit).Offset(offset).Find(&torrents).Error; err != nil {
		return nil, 0, err
	}

	return r.toMetaDataSlice(torrents), total, nil
}

func (r *GormRepository) InsertMetadata(md dhtcclient.Metadata) bool {
	if r.config.EnableBlacklist && r.IsBlacklisted(md) {
		log.Info().Msgf("Blacklisted: %s", md.Name)
		return false
	}

	filesJson, _ := json.Marshal(md.Files)
	torrent := GormTorrent{
		Name:         md.Name,
		InfoHash:     fmt.Sprintf("%x", md.InfoHash),
		Files:        string(filesJson),
		DiscoveredOn: md.DiscoveredOn,
		TotalSize:    md.TotalSize,
		Categories:   strings.Join(Categorize(md), ","),
	}

	err := r.db.Create(&torrent).Error
	return err == nil
}

func (r *GormRepository) GetWatchEntries() []WatchEntry {
	var entries []GormWatch
	r.db.Find(&entries)
	res := make([]WatchEntry, len(entries))
	for i, e := range entries {
		res[i] = WatchEntry{
			Id:        fmt.Sprint(e.ID),
			Key:       e.Key,
			MatchType: e.MatchType,
			Content:   e.Content,
		}
	}
	return res
}

func (r *GormRepository) InsertWatchEntry(key string, searchType string, searchInput string) bool {
	entry := GormWatch{
		Key:       key,
		MatchType: searchType,
		Content:   searchInput,
	}
	return r.db.Create(&entry).Error == nil
}

func (r *GormRepository) DeleteWatchEntry(id string) error {
	return r.db.Delete(&GormWatch{}, id).Error
}

func (r *GormRepository) GetBlacklistEntries() []BlacklistEntry {
	var entries []GormBlacklist
	r.db.Find(&entries)
	res := make([]BlacklistEntry, len(entries))
	for i, e := range entries {
		res[i] = BlacklistEntry{
			Id:     fmt.Sprint(e.ID),
			Filter: e.Filter,
			Type:   GetBlacklistTypeFromStrInt(e.Type),
		}
	}
	return res
}

func (r *GormRepository) AddToBlacklist(filters []string, entryType string) bool {
	for _, f := range filters {
		var count int64
		r.db.Model(&GormBlacklist{}).Where("filter = ? AND type = ?", f, entryType).Count(&count)
		if count == 0 {
			r.db.Create(&GormBlacklist{Filter: f, Type: entryType})
		}
	}
	return true
}

func (r *GormRepository) DeleteBlacklistItem(id string) error {
	return r.db.Delete(&GormBlacklist{}, id).Error
}

func (r *GormRepository) IsBlacklisted(md dhtcclient.Metadata) bool {
	var entries []GormBlacklist
	r.db.Find(&entries)

	for _, e := range entries {
		filter, err := regexp.Compile(e.Filter)
		if err != nil {
			continue
		}

		switch e.Type {
		case "0":
			if filter.MatchString(md.Name) {
				return true
			}
		case "1":
			if IsFileBlacklisted(md, filter) {
				return true
			}
		}
	}
	return false
}

func (r *GormRepository) toMetaDataSlice(torrents []GormTorrent) []MetaData {
	res := make([]MetaData, len(torrents))
	for i, t := range torrents {
		var files []interface{}
		json.Unmarshal([]byte(t.Files), &files)
		res[i] = MetaData{
			Name:         t.Name,
			InfoHash:     t.InfoHash,
			DiscoveredOn: time.Unix(t.DiscoveredOn, 0).Format(time.RFC822),
			TotalSize:    t.TotalSize,
			Files:        files,
			Categories:   strings.Split(t.Categories, ","),
		}
	}
	return res
}

func (r *GormRepository) GetAllInfoHashes() ([]string, error) {
	var hashes []string
	err := r.db.Model(&GormTorrent{}).Pluck("info_hash", &hashes).Error
	return hashes, err
}

func (r *GormRepository) GetStatsByInterval(interval string, limit int) ([]Stats, error) {
	duration := GetIntervalDuration(interval)
	seconds := int64(duration.Seconds())

	now := time.Now().Truncate(duration)
	since := now.Add(-time.Duration(limit) * duration).Unix()

	type Result struct {
		Bucket int64
		Count  int64
	}
	var results []Result

	bucketExpr := fmt.Sprintf("(discovered_on / %d) * %d", seconds, seconds)
	if r.db.Dialector.Name() == "mysql" {
		bucketExpr = fmt.Sprintf("(discovered_on DIV %d) * %d", seconds, seconds)
	}

	err := r.db.Model(&GormTorrent{}).
		Select(fmt.Sprintf("%s as bucket, count(*) as count", bucketExpr)).
		Where("discovered_on >= ?", since).
		Group("bucket").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	bucketMap := make(map[int64]int64)
	for _, res := range results {
		bucketMap[res.Bucket] = res.Count
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

func (r *GormRepository) InsertStats(stats Stats) error {
	return r.db.Create(&GormStats{
		Timestamp:    stats.Timestamp,
		TorrentCount: stats.TorrentCount,
	}).Error
}

func (r *GormRepository) GetCategoryDistribution() (map[string]int64, error) {
	type Result struct {
		Category string
		Count    int64
	}
	var results []Result
	err := r.db.Model(&GormTorrent{}).Select("category, count(*) as count").Group("category").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	dist := make(map[string]int64)
	for _, res := range results {
		cats := strings.Split(res.Category, ",")
		for _, cat := range cats {
			if cat == "" {
				cat = "unknown"
			}
			dist[cat] += res.Count
		}
	}
	return dist, nil
}

func (r *GormRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
