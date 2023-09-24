package db

import (
	"dhtc/config"
	dhtcclient "dhtc/dhtc-client"
	"encoding/hex"
	"fmt"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	"math/rand"
)

const WatchTable = "watches"
const BlacklistTable = "blacklist"
const TorrentTable = "torrents"

func GetInfoHashCount(db *clover.DB) int {
	vals, _ := db.FindAll(query.NewQuery(TorrentTable))
	return len(vals)
}

func FindBy(db *clover.DB, key string, searchType string, searchInput string) []MetaData {
	values, _ := db.FindAll(query.NewQuery(TorrentTable).MatchFunc(func(doc *document.Document) bool {
		return Matches(doc, key, searchType, searchInput)
	}))
	return Documents2MetaData(values)
}

func GetNRandomEntries(db *clover.DB, N int) []MetaData {
	count, _ := db.Count(query.NewQuery(TorrentTable))
	if count < N {
		N = count
	}

	all, _ := db.FindAll(query.NewQuery(TorrentTable))
	rVal := make([]*document.Document, N)
	for i := 0; i < N; i++ {
		rVal[i] = all[rand.Intn(count)]
	}
	return Documents2MetaData(rVal)
}

func InsertMetadata(config *config.Configuration, db *clover.DB, md dhtcclient.Metadata) bool {
	if config.EnableBlacklist && IsBlacklisted(db, md) {
		fmt.Printf("Blacklisted: %s\n", md.Name)
		return false
	}

	doc := document.NewDocument()
	doc.Set("Name", md.Name)
	doc.Set("InfoHash", hex.EncodeToString(md.InfoHash))
	doc.Set("Files", md.Files)
	doc.Set("DiscoveredOn", md.DiscoveredOn)
	doc.Set("TotalSize", md.TotalSize)
	_, err := db.InsertOne(TorrentTable, doc)
	if err != nil {
		return false
	} else {
		return true
	}
}

func OpenDatabase(config *config.Configuration) *clover.DB {
	db, err := clover.Open(config.DbName)

	if err != nil {
		fmt.Println("Error:", err)
	}

	_ = db.CreateCollection(TorrentTable)
	_ = db.CreateCollection(WatchTable)
	_ = db.CreateCollection(BlacklistTable)

	return db
}
