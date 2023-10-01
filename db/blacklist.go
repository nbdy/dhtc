package db

import (
	dhtcclient "dhtc/dhtc-client"
	"fmt"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	"github.com/rs/zerolog/log"
	"regexp"
)

func IsInBlacklist(db *clover.DB, filter string, entryType string) bool {
	result, _ := db.FindAll(query.NewQuery(BlacklistTable).Where(query.Field("Filter").Eq(filter).And(query.Field("Type").Eq(entryType))))
	return len(result) > 0
}

func AddToBlacklist(db *clover.DB, filters []string, entryType string) bool {
	rVal := false

	for _, filter := range filters {
		if !IsInBlacklist(db, filter, entryType) {
			fmt.Printf("Is not in blacklist yet: %s\n", filter)
			doc := document.NewDocument()
			doc.Set("Type", entryType)
			doc.Set("Filter", filter)
			_, err := db.InsertOne(BlacklistTable, doc)
			if err != nil {
				log.Error().Err(err)
			}
			rVal = err == nil
		}
	}

	return rVal
}

func IsFileBlacklisted(md dhtcclient.Metadata, filter *regexp.Regexp) bool {
	rVal := false
	for i := 0; i < len(md.Files); i++ {
		rVal = filter.MatchString(md.Files[i].Path)
		if rVal {
			break
		}
	}
	return rVal
}

func IsBlacklisted(db *clover.DB, md dhtcclient.Metadata) bool {
	all, _ := db.FindAll(query.NewQuery(BlacklistTable).MatchFunc(func(doc *document.Document) bool {
		filterStr := doc.Get("Filter").(string)
		filter := regexp.MustCompile(filterStr)

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

func DeleteBlacklistItem(db *clover.DB, itemId string) bool {
	fmt.Println("Removing", itemId)
	return db.DeleteById(BlacklistTable, itemId) == nil
}

func GetBlacklistTypeFromStrInt(entryType string) string {
	switch entryType {
	case "0":
		return "Name"
	case "1":
		return "File name"
	}
	return "Unknown"
}

func GetBlacklistEntries(db *clover.DB) []BlacklistEntry {
	all, _ := db.FindAll(query.NewQuery(BlacklistTable))
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
