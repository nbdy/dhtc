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

	"github.com/ostafen/clover/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	telegram "gopkg.in/telebot.v3"
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

func crawl(configuration *config.Configuration, bootstrapNodes []string, database *clover.DB, bot *telegram.Bot) {
	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	trawlingManager := dhtcclient.NewManager(bootstrapNodes, indexerAddrs, 1, configuration.MaxNeighbors)
	metadataSink := dhtcclient.NewSink(configuration.DrainTimeout, configuration.MaxLeeches)

	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			hash := result.InfoHash()

			if !cache.InfoHashCache.Contains(hash) {
				cache.InfoHashCache.Add(hash)
				metadataSink.Sink(result)
			}

		case md := <-metadataSink.Drain():
			if db.InsertMetadata(configuration, database, md) {
				fmt.Println("\t + Added:", md.Name)
				db.CheckWatches(configuration, database, md, bot)
			}

		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.ParseArguments()
	database := db.OpenDatabase(cfg)

	cache.PopulateInfoHashCacheFromDatabase(database)

	db.AddToBlacklist(database, ReadFileLines(cfg.NameBlacklist), "0")
	db.AddToBlacklist(database, ReadFileLines(cfg.FileBlacklist), "1")

	bootstrapNodes := ReadFileLines(cfg.BootstrapNodeFile)

	if len(bootstrapNodes) == 0 {
		log.Warn().Msg("No bootstrap nodes found in '" + cfg.BootstrapNodeFile + "'.")
		log.Info().Msg("Using default bootstrap nodes.")
		bootstrapNodes = []string{
			"router.bittorrent.com:6881", "router.utorrent.com:6881",
			"dht.transmissionbt.com:6881", "dht.libtorrent.org:25401",
		}
	}

	if !cfg.OnlyWebServer {
		bot := notifier.SetupTelegramBot(cfg)

		for i := 0; i < cfg.CrawlerThreads; i++ {
			go crawl(cfg, bootstrapNodes, database, bot)
		}
	}

	ui.RunWebServer(cfg, database)
}
