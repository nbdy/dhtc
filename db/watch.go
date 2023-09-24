package db

import (
	"dhtc/config"
	dhtcclient "dhtc/dhtc-client"
	"dhtc/notifier"
	"fmt"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	telegram "gopkg.in/telebot.v3"
)

func GetWatchEntries(db *clover.DB) []WatchEntry {
	all, _ := db.FindAll(query.NewQuery(WatchTable))
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

func CheckWatches(config *config.Configuration, db *clover.DB, md dhtcclient.Metadata, bot *telegram.Bot) {
	if bot == nil {
		return
	}

	all, _ := db.FindAll(query.NewQuery(WatchTable))
	for _, d := range all {
		key := d.Get("Key").(string)
		matchType := d.Get("MatchType").(string)
		content := d.Get("").(string)
		if len(FindBy(db, key, matchType, content)) > 0 {
			msg := ""
			if matchType == "Files" {
				msg = fmt.Sprintf("Match found: '%s' contains file which %s '%s'.", md.Name, matchType, content)
			} else {
				msg = fmt.Sprintf("Match found: '%s' %s '%s'", md.Name, matchType, content)
			}
			notifier.NotifyTelegram(config, bot, msg)
		}
	}
}

func InsertWatchEntry(db *clover.DB, key string, searchType string, searchInput string) bool {
	doc := document.NewDocument()
	doc.Set("Key", key)
	doc.Set("MatchType", searchType)
	doc.Set("Content", searchInput)
	_, err := db.InsertOne(WatchTable, doc)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
}

func DeleteWatchEntry(db *clover.DB, entryId string) bool {
	return db.DeleteById(WatchTable, entryId) == nil
}
