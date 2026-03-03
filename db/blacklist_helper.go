package db

import (
	dhtcclient "dhtc/dhtc-client"
	"regexp"
)

func IsFileBlacklisted(md dhtcclient.Metadata, filter *regexp.Regexp) bool {
	for i := range len(md.Files) {
		if filter.MatchString(md.Files[i].Path) {
			return true
		}
	}
	return false
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
