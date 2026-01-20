package main

import (
	"bufio"
	"dhtc/cache"
	"dhtc/config"
	"dhtc/db"
	dhtcclient "dhtc/dhtc-client"
	"dhtc/notifier"
	"dhtc/ui"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ReadFileLines(filePath string) []string {
	var rVal []string

	file, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err)
		return rVal
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rVal = append(rVal, scanner.Text())
	}

	return rVal
}

func crawl(configuration *config.Configuration, bootstrapNodes []string, database db.Repository, nManager *notifier.Manager, hub *ui.Hub) {
	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	trawlingManager := dhtcclient.NewManager(bootstrapNodes, indexerAddrs, 10*time.Second, configuration.MaxNeighbors, configuration.RateLimit)
	metadataSink := dhtcclient.NewSink(configuration.DrainTimeout, configuration.MaxLeeches, configuration.MaxConcurrentDownloads)

	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			hash := result.InfoHash()

			if !cache.InfoHashCache.Contains(string(hash)) {
				cache.InfoHashCache.Add(string(hash))
				metadataSink.Sink(result)
			}

		case md := <-metadataSink.Drain():
			if database.InsertMetadata(md) {
				fmt.Println("\t + Added:", md.Name)
				db.CheckWatches(configuration, database, md, nManager)
				hub.BroadcastMetadata(md)
			}

		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}

func collectStats(database db.Repository) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		count := database.GetInfoHashCount()
		err := database.InsertStats(db.Stats{
			Timestamp:    time.Now().Truncate(time.Minute),
			TorrentCount: int64(count),
		})
		if err != nil {
			log.Error().Err(err).Msg("could not insert stats")
		}
		<-ticker.C
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.ParseArguments()
	database, err := db.OpenRepository(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("could not open database")
	}

	cache.PopulateInfoHashCacheFromDatabase(database)

	database.AddToBlacklist(ReadFileLines(cfg.NameBlacklist), "0")
	database.AddToBlacklist(ReadFileLines(cfg.FileBlacklist), "1")

	bootstrapNodes := ReadFileLines(cfg.BootstrapNodeFile)

	if len(bootstrapNodes) == 0 {
		log.Warn().Msg("No bootstrap nodes found in '" + cfg.BootstrapNodeFile + "'.")
		log.Info().Msg("Using default bootstrap nodes.")
		bootstrapNodes = []string{
			"router.bittorrent.com:6881", "router.utorrent.com:6881",
			"dht.transmissionbt.com:6881", "dht.libtorrent.org:25401",
		}
	}

	hub := ui.NewHub()
	go hub.Run()

	var nManager *notifier.Manager
	if !cfg.OnlyWebServer {
		nManager = notifier.SetupNotifiers(cfg)

		if cfg.Statistics {
			go collectStats(database)
		}

		for i := 0; i < cfg.CrawlerThreads; i++ {
			go crawl(cfg, bootstrapNodes, database, nManager, hub)
		}
	}

	ui.RunWebServer(cfg, database, hub, nManager)
}
