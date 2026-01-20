package cache

import (
	"dhtc/db"
	"encoding/hex"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"
)

var InfoHashCache = mapset.NewSet[string]()

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
		InfoHashCache.Add(string(h))
	}

	log.Debug().Msgf("info hash cache size %d elements", InfoHashCache.Cardinality())
}
