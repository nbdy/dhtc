package cache

import (
	"dhtc/db"
	"encoding/hex"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/query"
	"github.com/rs/zerolog/log"
)

var InfoHashCache = mapset.NewSet[[20]byte]()

func PopulateInfoHashCacheFromDatabase(database clover.DB) {
	all, _ := database.FindAll(query.NewQuery(db.TorrentTable))
	for _, d := range all {
		ih := d.Get("InfoHash").(string)
		h, err := hex.DecodeString(ih)
		if len(h) != 20 || err != nil {
			fmt.Print("x")
			continue
		}
		h20 := (*[20]byte)(h)
		InfoHashCache.Add(*h20)
	}

	log.Debug().Msgf("info hash cache size %d elements\n", InfoHashCache.Cardinality())
}
