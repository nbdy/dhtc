package main

import (
	"embed"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/boramalper/magnetico/cmd/magneticod/bittorrent/metadata"
	"github.com/boramalper/magnetico/cmd/magneticod/dht"
	"github.com/labstack/echo/v4"
	"github.com/noirbizarre/gonja"
	"github.com/ostafen/clover"
	telegram "gopkg.in/telebot.v3"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type MetaData struct {
	Name         string
	InfoHash     string
	DiscoveredOn string
	TotalSize    uint64
	Files        []interface{}
}

type Configuration struct {
	dbName       string
	address      string
	maxNeighbors uint
	maxLeeches   int
	drainTimeout time.Duration

	telegramToken    string
	telegramUsername string

	safeMode bool
}

type WatchEntry struct {
	Id        string
	Key       string
	MatchType string
	Content   string
}

//go:embed templates
var templates embed.FS

//go:embed static
var static embed.FS

var db *clover.DB
var config = Configuration{}
var bot *telegram.Bot = nil

func notifyTelegram(message string) {
	if bot != nil {
		chat, err := bot.ChatByUsername(config.telegramUsername)
		if err == nil {
			fmt.Println("Sending telegram notification.")
			_, err := bot.Send(chat, message)
			if err != nil {
				fmt.Printf("Could not send message '%s' to user '%s'.", message, config.telegramUsername)
			}
		} else {
			fmt.Printf("Could not find chat by username '%s'\n", config.telegramUsername)
		}
	}
}

func checkWatches(md metadata.Metadata) {
	if bot == nil {
		return
	}

	all, _ := db.Query("watches").FindAll()
	for _, document := range all {
		key := document.Get("Key").(string)
		matchType := document.Get("MatchType").(string)
		content := document.Get("").(string)
		if len(findBy(key, matchType, content)) > 0 {
			msg := ""
			if matchType == "Files" {
				msg = fmt.Sprintf("Match found: '%s' contains file which %s '%s'.", md.Name, matchType, content)
			} else {
				msg = fmt.Sprintf("Match found: '%s' %s '%s'", md.Name, matchType, content)
			}
			notifyTelegram(msg)
		}
	}
}

func hasInfoHash(InfoHash string) bool {
	values, err := db.Query("torrents").Where(clover.Field("InfoHash").Eq(InfoHash)).FindAll()
	if err != nil {
		return false
	}
	return len(values) > 0
}

func getInfoHashCount() int {
	vals, _ := db.Query("torrents").FindAll()
	return len(vals)
}

func document2MetaData(value *clover.Document) MetaData {
	return MetaData{
		value.Get("Name").(string),
		value.Get("InfoHash").(string),
		time.Unix(int64(value.Get("DiscoveredOn").(float64)), 0).Format(time.RFC822),
		uint64(value.Get("TotalSize").(float64)),
		value.Get("Files").([]interface{}),
	}
}

func documents2MetaData(values []*clover.Document) []MetaData {
	rVal := make([]MetaData, len(values))
	for i, value := range values {
		rVal[i] = document2MetaData(value)
	}
	return rVal
}

func matchString(searchType string, x string, y string) bool {
	rVal := false
	switch searchType {
	case "contains":
		rVal = strings.Contains(strings.ToLower(x), strings.ToLower(y))
	case "equals":
		rVal = strings.ToLower(x) == strings.ToLower(y)
	case "startswith":
		rVal = strings.HasPrefix(strings.ToLower(x), strings.ToLower(y))
	case "endswith":
		rVal = strings.HasSuffix(strings.ToLower(x), strings.ToLower(y))
	default:
		rVal = false
	}
	return rVal
}

func hasMatchingFile(fileList []interface{}, searchType string, searchInput string) bool {
	rVal := false
	for _, item := range fileList {
		for key, value := range item.(map[string]interface{}) {
			if key == "path" {
				rVal = matchString(searchType, value.(string), searchInput)
			}
		}
	}
	return rVal
}

func foundOnDate(date time.Time, searchInput string) bool {
	parsedInput, err := time.Parse("2006.01.02", searchInput)
	if err != nil {
		fmt.Println(err)
		return false
	}
	endDate := parsedInput.AddDate(0, 0, 1)
	return date.After(parsedInput) && date.Before(endDate)
}

func matches(doc *clover.Document, key string, searchType string, searchInput string) bool {
	rVal := doc.Has(key)
	if rVal {
		value := doc.Get(key)

		if key == "Files" {
			rVal = hasMatchingFile(value.([]interface{}), searchType, searchInput)
		} else if key == "DiscoveredOn" {
			rVal = foundOnDate(time.Unix(int64(value.(float64)), 0), searchInput)
		} else {
			rVal = matchString(searchType, value.(string), searchInput)
		}
	}
	return rVal
}

func findBy(key string, searchType string, searchInput string) []MetaData {
	values, _ := db.Query("torrents").MatchPredicate(func(doc *clover.Document) bool {
		return matches(doc, key, searchType, searchInput)
	}).FindAll()
	return documents2MetaData(values)
}

func getNRandomEntries(N int) []MetaData {
	rand.Seed(time.Now().Unix())
	count, _ := db.Query("torrents").Count()
	if count < N {
		N = count
	}
	all, _ := db.Query("torrents").FindAll()
	rVal := make([]*clover.Document, N)
	for i := 0; i < N; i++ {
		rVal[i] = all[rand.Intn(count)]
	}
	return documents2MetaData(rVal)
}

func getWatchEntries() []WatchEntry {
	all, _ := db.Query("watches").FindAll()
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

func insertMetadata(md metadata.Metadata) bool {
	doc := clover.NewDocument()
	doc.Set("Name", md.Name)
	doc.Set("InfoHash", hex.EncodeToString(md.InfoHash))
	doc.Set("Files", md.Files)
	doc.Set("DiscoveredOn", md.DiscoveredOn)
	doc.Set("TotalSize", md.TotalSize)
	_, err := db.InsertOne("torrents", doc)
	if err != nil {
		return false
	} else {
		return true
	}
}

func insertWatchEntry(key string, searchType string, searchInput string) bool {
	doc := clover.NewDocument()
	doc.Set("Key", key)
	doc.Set("MatchType", searchType)
	doc.Set("Content", searchInput)
	_, err := db.InsertOne("watches", doc)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
}

func deleteWatchEntry(entryId string) bool {
	err := db.Query("watches").DeleteById(entryId)
	if err != nil {
		return false
	}
	return true
}

func crawl() {
	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	trawlingManager := dht.NewManager(indexerAddrs, 1, config.maxNeighbors)
	metadataSink := metadata.NewSink(config.drainTimeout, config.maxLeeches)

	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			hash := result.InfoHash()
			if !hasInfoHash(hex.EncodeToString(hash[:])) {
				metadataSink.Sink(result)
			}

		case md := <-metadataSink.Drain():
			if insertMetadata(md) {
				fmt.Println("\t + Added:", md.Name)
				checkWatches(md)
			}

		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}

var dashboardTplBytes, _ = templates.ReadFile("templates/dashboard.html")
var dashboardTpl = gonja.Must(gonja.FromBytes(dashboardTplBytes))

func dashboard(c echo.Context) error {
	out, _ := dashboardTpl.Execute(gonja.Context{"info_hash_count": getInfoHashCount(), "path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

var searchTplBytes, _ = templates.ReadFile("templates/search.html")
var searchTpl = gonja.Must(gonja.FromBytes(searchTplBytes))

func searchGet(c echo.Context) error {
	out, _ := searchTpl.Execute(gonja.Context{"path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

func searchPost(c echo.Context) error {
	out, _ := searchTpl.Execute(gonja.Context{"results": findBy(c.FormValue("key"), c.FormValue("match-type"), c.FormValue("search-input")), "path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

var discoverTplBytes, _ = templates.ReadFile("templates/discover.html")
var discoverTpl = gonja.Must(gonja.FromBytes(discoverTplBytes))

func discoverGet(c echo.Context) error {
	out, _ := discoverTpl.Execute(gonja.Context{"results": getNRandomEntries(50), "path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

func discoverPost(c echo.Context) error {
	N, err := strconv.Atoi(c.FormValue("limit"))
	if err != nil {
		N = 50
	}
	out, _ := discoverTpl.Execute(gonja.Context{"results": getNRandomEntries(N), "path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

var watchTplBytes, _ = templates.ReadFile("templates/watches.html")
var watchTpl = gonja.Must(gonja.FromBytes(watchTplBytes))

func watchGet(c echo.Context) error {
	out, _ := watchTpl.Execute(gonja.Context{"path": c.Path(), "results": getWatchEntries()})
	return c.HTML(http.StatusOK, out)
}

func watchPost(c echo.Context) error {
	opOk := false
	op := c.FormValue("op")
	fmt.Println("Operation:", op)
	if op == "add" {
		opOk = insertWatchEntry(c.FormValue("key"), c.FormValue("match-type"), c.FormValue("search-input"))
	} else if op == "delete" {
		opOk = deleteWatchEntry(c.FormValue("id"))
	}
	out, _ := watchTpl.Execute(gonja.Context{"path": c.Path(), "op": op, "opOk": opOk, "results": getWatchEntries()})
	return c.HTML(http.StatusOK, out)
}

func webserver() {
	srv := echo.New()

	srv.GET("", dashboard)
	srv.GET("/dashboard", dashboard)
	srv.GET("/search", searchGet)
	srv.POST("/search", searchPost)
	srv.GET("/discover", discoverGet)
	srv.POST("/discover", discoverPost)
	srv.GET("/watches", watchGet)
	srv.POST("/watches", watchPost)

	srv.StaticFS("/css", echo.MustSubFS(static, "static/css"))
	srv.StaticFS("/js", echo.MustSubFS(static, "static/js"))

	err := srv.Start(config.address)
	if err != nil {
		return
	}
}

func parseArguments() {
	flag.StringVar(&config.dbName, "database", "dhtdb", "database name")
	flag.StringVar(&config.address, "address", ":4200", "address to run on")
	flag.UintVar(&config.maxNeighbors, "maxNeighbors", 1000, "max. indexer neighbors")
	flag.IntVar(&config.maxLeeches, "maxLeeches", 256, "max. leeches")
	flag.DurationVar(&config.drainTimeout, "drainTimeout", 5*time.Second, "drain timeout")

	flag.StringVar(&config.telegramToken, "telegramToken", "", "bot token for notifications")
	flag.StringVar(&config.telegramUsername, "telegramUsername", "", "username to send notifications to")

	flag.BoolVar(&config.safeMode, "safeMode", false, "start with safe mode enabled")

	flag.Parse()
}

func openDatabase() {
	var err error
	db, err = clover.Open(config.dbName)
	if err != nil {
		fmt.Println("Error:", err)
	}

	err = db.CreateCollection("torrents")
	if err != nil {
		fmt.Println("Error:", err)
	}

	err = db.CreateCollection("watches")
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func setupTelegramBot() {
	if config.telegramToken != "" {
		pref := telegram.Settings{
			Token:  config.telegramToken,
			Poller: &telegram.LongPoller{Timeout: 10 * time.Second},
		}

		var err error
		bot, err = telegram.NewBot(pref)
		if err != nil {
			fmt.Println("Could not create telegram bot.")
		}
	}
}

func main() {
	parseArguments()
	openDatabase()
	setupTelegramBot()

	go crawl()
	webserver()
}
