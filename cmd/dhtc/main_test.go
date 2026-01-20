package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSystemStartup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	os.Args = []string{"dhtc", "-OnlyWebServer", "-address", "127.0.0.1:4201"}

	go main()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("System did not come up within 10 seconds")
		case <-ticker.C:
			resp, err := http.Get("http://127.0.0.1:4201/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				fmt.Println("System is up!")
				return
			}
		}
	}
}
