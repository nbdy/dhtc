package dhtc_client

import (
	"net"
	"sync"
	"time"
)

type mDict struct {
	UTMetadata int `bencode:"ut_metadata"`
	UTPex      int `bencode:"ut_pex"`
}

type pexMsg struct {
	Added   string `bencode:"added"`
	AddedF  string `bencode:"added.f"`
	Added6  string `bencode:"added6"`
	Added6F string `bencode:"added6.f"`
}

type rootDict struct {
	M            mDict `bencode:"m"`
	MetadataSize int   `bencode:"metadata_size"`
}

type extDict struct {
	MsgType int `bencode:"msg_type"`
	Piece   int `bencode:"piece"`
}

type Statistics struct {
	NDiscovered map[string]uint64 `json:"nDiscovered"`
	NFiles      map[string]uint64 `json:"nFiles"`
	TotalSize   map[string]uint64 `json:"totalSize"`

	// All these slices below have the exact length equal to the Period.
	//NDiscovered []uint64  `json:"nDiscovered"`

}

type File struct {
	Size int64  `json:"size"`
	Path string `json:"path"`
}

type TorrentMetadata struct {
	ID           uint64  `json:"id"`
	InfoHash     []byte  `json:"infoHash"` // marshalled differently
	Name         string  `json:"name"`
	Size         uint64  `json:"size"`
	DiscoveredOn int64   `json:"discoveredOn"`
	NFiles       uint    `json:"nFiles"`
	Relevance    float64 `json:"relevance"`
}

type SimpleTorrentSummary struct {
	InfoHash string `json:"infoHash"`
	Name     string `json:"name"`
	Files    []File `json:"files"`
}

type Metadata struct {
	InfoHash []byte
	// Name should be thought of "Title" of the torrent. For single-file torrents, it is the name
	// of the file, and for multi-file torrents, it is the name of the root directory.
	Name         string
	TotalSize    uint64
	DiscoveredOn int64
	// Files must be populated for both single-file and multi-file torrents!
	Files []File
}

type Sink struct {
	PeerID                 []byte
	deadline               time.Duration
	maxNLeeches            int
	maxConcurrentDownloads int
	downloadSem            chan struct{}
	drain                  chan Metadata

	incomingInfoHashes   map[string][]net.TCPAddr
	incomingInfoHashesMx sync.Mutex

	terminated  bool
	termination chan interface{}

	deleted int
}
