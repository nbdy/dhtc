package main

import (
	"fmt"
	"github.com/boramalper/magnetico/cmd/magneticod/bittorrent/metadata"
	"github.com/boramalper/magnetico/cmd/magneticod/dht"
	"github.com/labstack/echo/v4"
	"github.com/noirbizarre/gonja"
	"github.com/ostafen/clover"
	"net/http"
	"os"
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

func findByName(name string) []MetaData {
	values, _ := db.Query("torrents").MatchPredicate(func(doc *clover.Document) bool {
		return strings.Contains(strings.ToLower(doc.Get("Name").(string)), strings.ToLower(name))
	}).FindAll()
	return document2MetaData(values)
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
	out, _ := dashboardTpl.Execute(gonja.Context{"info_hash_count": getInfoHashCount()})
	return c.HTML(http.StatusOK, out)
}

var searchTpl, _ = gonja.FromFile("templates/search.html")

func searchGet(c echo.Context) error {
	out, _ := searchTpl.Execute(gonja.Context{})
	return c.HTML(http.StatusOK, out)
}

func searchPost(c echo.Context) error {
	out, _ := searchTpl.Execute(gonja.Context{"results": findByName(c.FormValue("searchInput"))})
	return c.HTML(http.StatusOK, out)
}

func webserver() {
	srv := echo.New()
	srv.GET("", dashboard)
	srv.GET("dashboard", dashboard)
	srv.GET("search", searchGet)
	srv.POST("search", searchPost)

	err := srv.Start(":4200")
	if err != nil {
		return
	}
}

func main() {
	go crawl()
	webserver()
}
