package main

import (
	"fmt"
	"github.com/boramalper/magnetico/cmd/magneticod/bittorrent/metadata"
	"github.com/boramalper/magnetico/cmd/magneticod/dht"
	"github.com/labstack/echo/v4"
	"github.com/noirbizarre/gonja"
	"github.com/ostafen/clover"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var db, _ = clover.Open("dhtdb")

func hasInfoHash(InfoHash [20]byte) bool {
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

type MetaData struct {
	Name         string
	DiscoveredOn string
	TotalSize    uint64
	FileCount    uint64
}

func document2MetaData(values []*clover.Document) []MetaData {
	rVal := make([]MetaData, len(values))
	for i, value := range values {
		rVal[i] = MetaData{
			value.Get("Name").(string),
			time.Unix(int64(value.Get("DiscoveredOn").(float64)), 0).Format(time.RFC822),
			uint64(value.Get("TotalSize").(float64)),
			uint64(len(value.Get("Files").([]interface{}))),
		}
	}
	return rVal
}

func matchString(searchType string, x string, y string) bool {
	rVal := false
	switch searchType {
	case "0":
		rVal = strings.Contains(strings.ToLower(x), strings.ToLower(y))
	case "1":
		rVal = strings.ToLower(x) == strings.ToLower(y)
	case "2":
		rVal = strings.HasPrefix(strings.ToLower(x), strings.ToLower(y))
	case "3":
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
			if key == "Path" {
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
	dbKey := "Name"

	switch key {
	case "0":
		dbKey = "Name"
	case "1":
		dbKey = "InfoHash"
	case "2":
		dbKey = "Files"
	case "3":
		dbKey = "DiscoveredOn"
	}

	rVal := doc.Has(dbKey)
	if rVal {
		value := doc.Get(dbKey)

		if key == "2" {
			rVal = hasMatchingFile(value.([]interface{}), searchType, searchInput)
		} else if key == "3" {
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
	fmt.Println("Found", len(values), "results")
	return document2MetaData(values)
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
		rVal[i] = all[rand.Intn(N)]
	}
	return document2MetaData(rVal)
}

func insertMetadata(md metadata.Metadata) bool {
	doc := clover.NewDocument()
	doc.Set("Name", md.Name)
	doc.Set("InfoHash", md.InfoHash)
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

func crawl() {
	err := db.CreateCollection("torrents")
	if err != nil {
		fmt.Println("Error:", err)
	}

	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	trawlingManager := dht.NewManager(indexerAddrs, 1, 1000)
	metadataSink := metadata.NewSink(5*time.Second, 128)

	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			if !hasInfoHash(result.InfoHash()) {
				metadataSink.Sink(result)
			}

		case md := <-metadataSink.Drain():
			if insertMetadata(md) {
				fmt.Println("\t + Added:", md.Name)
			}

		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}

var dashboardTpl, _ = gonja.FromFile("templates/dashboard.html")

func dashboard(c echo.Context) error {
	out, _ := dashboardTpl.Execute(gonja.Context{"info_hash_count": getInfoHashCount(), "path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

var searchTpl, _ = gonja.FromFile("templates/search.html")

func searchGet(c echo.Context) error {
	out, _ := searchTpl.Execute(gonja.Context{"path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

func searchPost(c echo.Context) error {
	out, _ := searchTpl.Execute(gonja.Context{"results": findBy(c.FormValue("key"), c.FormValue("match-type"), c.FormValue("search-input")), "path": c.Path()})
	return c.HTML(http.StatusOK, out)
}

var discoverTpl, _ = gonja.FromFile("templates/discover.html")

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

func webserver() {
	srv := echo.New()
	srv.GET("", dashboard)
	srv.GET("dashboard", dashboard)
	srv.GET("search", searchGet)
	srv.POST("search", searchPost)
	srv.GET("discover", discoverGet)
	srv.POST("discover", discoverPost)

	err := srv.Start(":4200")
	if err != nil {
		return
	}
}

func main() {
	go crawl()
	webserver()
}
