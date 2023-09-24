package main

import (
	"dhtc/cache"
	"dhtc/config"
	"dhtc/db"
	dhtcclient "dhtc/dhtc-client"
	"dhtc/notifier"
	"dhtc/ui"
	"fmt"
	"github.com/ostafen/clover/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	telegram "gopkg.in/telebot.v3"
	"os"
	"strings"
)

func ReadFileLines(filePath string) []string {
	var rVal []string

	content, err := os.ReadFile(filePath)
	if err == nil {
		rVal = strings.Split(string(content), "\n")
	} else {
		log.Error().Err(err)
	}

	return rVal
}

func crawl(configuration *config.Configuration, database *clover.DB, bot *telegram.Bot) {
	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	bootstrapNodes := ReadFileLines(configuration.BootstrapNodeFile)

	if len(bootstrapNodes) == 0 {
		log.Error().Msg("No bootstrap nodes")
		return
	}

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

	db.AddToBlacklist(database, ReadFileLines(cfg.NameBlacklist), "0")
	db.AddToBlacklist(database, ReadFileLines(cfg.FileBlacklist), "1")

	bot := notifier.SetupTelegramBot(cfg)

	for i := 0; i < cfg.CrawlerThreads; i++ {
		go crawl(cfg, database, bot)
	}

	ui.RunWebServer(cfg, database)
}
