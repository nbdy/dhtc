package dhtc_client

import (
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

func NewSink(deadline time.Duration, maxNLeeches int, maxConcurrentDownloads int) *Sink {
	ms := new(Sink)

	ms.PeerID = randomID()
	ms.deadline = deadline
	ms.maxNLeeches = maxNLeeches
	ms.maxConcurrentDownloads = maxConcurrentDownloads
	ms.downloadSem = make(chan struct{}, maxConcurrentDownloads)
	ms.drain = make(chan Metadata, 10)
	ms.incomingInfoHashes = make(map[string][]net.TCPAddr)
	ms.termination = make(chan interface{})

	return ms
}

func (ms *Sink) Sink(res Result) {
	if ms.terminated {
		log.Panic().Msg("Trying to Sink() an already closed Sink!")
	}
	ms.incomingInfoHashesMx.Lock()
	defer ms.incomingInfoHashesMx.Unlock()

	// cap the max # of leeches
	if len(ms.incomingInfoHashes) >= ms.maxNLeeches {
		return
	}

	infoHash := res.InfoHash()
	peerAddrs := res.PeerAddrs()

	if _, exists := ms.incomingInfoHashes[string(infoHash)]; exists {
		return
	} else if len(peerAddrs) > 0 {
		peer := peerAddrs[0]
		ms.incomingInfoHashes[string(infoHash)] = peerAddrs[1:]

		go ms.download(infoHash, peer)
	}
}

func (ms *Sink) download(infoHash []byte, peer net.TCPAddr) {
	ms.downloadSem <- struct{}{}
	defer func() { <-ms.downloadSem }()

	NewClient(infoHash, &peer, ms.PeerID, ClientEventHandlers{
		OnSuccess: ms.flush,
		OnError:   ms.onLeechError,
		OnPeers:   ms.onPeers,
	}).Do(time.Now().Add(ms.deadline))
}

func (ms *Sink) onPeers(infoHash []byte, peers []net.TCPAddr) {
	ms.incomingInfoHashesMx.Lock()
	defer ms.incomingInfoHashesMx.Unlock()

	if _, exists := ms.incomingInfoHashes[string(infoHash)]; !exists {
		return
	}

	ms.incomingInfoHashes[string(infoHash)] = append(ms.incomingInfoHashes[string(infoHash)], peers...)
}

func (ms *Sink) Drain() <-chan Metadata {
	if ms.terminated {
		log.Panic().Msg("Trying to Drain() an already closed Sink!")
	}
	return ms.drain
}

func (ms *Sink) Terminate() {
	ms.terminated = true
	close(ms.termination)
	close(ms.drain)
}

func (ms *Sink) flush(result Metadata) {
	if ms.terminated {
		return
	}

	ms.drain <- result
	// Delete the infoHash from ms.incomingInfoHashes ONLY AFTER once we've flushed the
	// metadata!
	ms.incomingInfoHashesMx.Lock()
	defer ms.incomingInfoHashesMx.Unlock()

	delete(ms.incomingInfoHashes, string(result.InfoHash))
}

func (ms *Sink) onLeechError(infoHash []byte, err error) {
	log.Debug().Err(err)

	ms.incomingInfoHashesMx.Lock()
	defer ms.incomingInfoHashesMx.Unlock()

	if len(ms.incomingInfoHashes[string(infoHash)]) > 0 {
		peer := ms.incomingInfoHashes[string(infoHash)][0]
		ms.incomingInfoHashes[string(infoHash)] = ms.incomingInfoHashes[string(infoHash)][1:]
		go ms.download(infoHash, peer)
	} else {
		ms.deleted++
		delete(ms.incomingInfoHashes, string(infoHash))
	}
}
