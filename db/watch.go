package db

import (
	"dhtc/config"
	dhtcclient "dhtc/dhtc-client"
	"dhtc/notifier"
	"fmt"
)

func CheckWatches(config *config.Configuration, database Repository, md dhtcclient.Metadata, nManager *notifier.Manager) {
	if nManager == nil {
		return
	}

	all := database.GetWatchEntries()
	for _, d := range all {
		key := d.Key
		matchType := d.MatchType
		content := d.Content
		if len(database.FindBy(key, matchType, content)) > 0 {
			msg := ""
			if matchType == "Files" {
				msg = fmt.Sprintf("Match found: '%s' contains file which %s '%s'.", md.Name, matchType, content)
			} else {
				msg = fmt.Sprintf("Match found: '%s' %s '%s'", md.Name, matchType, content)
			}
			nManager.Notify(msg)
		}
	}
}
