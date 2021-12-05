package main

import (
	"context"
	"fmt"
	"github.com/boramalper/magnetico/cmd/magneticod/bittorrent/metadata"
	"github.com/boramalper/magnetico/cmd/magneticod/dht"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

func hasInfoHash(collection *mongo.Collection, InfoHash [20]byte) bool {
	c, _ := collection.CountDocuments(context.TODO(), bson.D{{"InfoHash", InfoHash}})
	return c != 0
}

func insertMetadata(collection *mongo.Collection, md metadata.Metadata) bool {
	_, e := collection.InsertOne(context.TODO(), md)

	return e == nil
}

func main() {
	mdbc, e := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if e != nil {
		panic(e)
	}
	defer func() {
		if e = mdbc.Disconnect(context.TODO()); e != nil {
			panic(e)
		}
	}()

	tcol := mdbc.Database("torrent").Collection("metadata")

	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	trawlingManager := dht.NewManager(indexerAddrs, 1, 1000)
	metadataSink := metadata.NewSink(5*time.Second, 128)

	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			if !hasInfoHash(tcol, result.InfoHash()) {
				metadataSink.Sink(result)
			}

		case md := <-metadataSink.Drain():
			if insertMetadata(tcol, md) {
				fmt.Println("Added:", md.Name)
			} else {
				fmt.Println("Error:", e)
			}

		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}
