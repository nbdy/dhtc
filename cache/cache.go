package cache

import (
	"dhtc/db"
	"encoding/hex"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"
)

var (
	infoHashCache   = mapset.NewSet[string]()
	infoHashCacheMu sync.RWMutex
)

func InfoHashCacheContains(infoHash string) bool {
	infoHashCacheMu.RLock()
	defer infoHashCacheMu.RUnlock()
	return infoHashCache.Contains(infoHash)
}

func InfoHashCacheAdd(infoHash string) bool {
	infoHashCacheMu.Lock()
	defer infoHashCacheMu.Unlock()
	return infoHashCache.Add(infoHash)
}

func PopulateInfoHashCacheFromDatabase(database db.Repository) {
	all, err := database.GetAllInfoHashes()
	if err != nil {
		log.Error().Err(err).Msg("could not get all info hashes from database")
		return
	}
	for _, ih := range all {
		h, err := hex.DecodeString(ih)
		if err != nil {
			continue
		}
		InfoHashCacheAdd(string(h))
	}

	infoHashCacheMu.RLock()
	cacheSize := infoHashCache.Cardinality()
	infoHashCacheMu.RUnlock()
	log.Debug().Msgf("info hash cache size %d elements", cacheSize)
}
