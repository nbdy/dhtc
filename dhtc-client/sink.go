package dhtc_client

import (
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

func NewSink(deadline time.Duration, maxNLeeches int) *Sink {
	ms := new(Sink)

	ms.PeerID = randomID()
	ms.deadline = deadline
	ms.maxNLeeches = maxNLeeches
	ms.drain = make(chan Metadata, 10)
	ms.incomingInfoHashes = make(map[[20]byte][]net.TCPAddr)
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

	if _, exists := ms.incomingInfoHashes[infoHash]; exists {
		return
	} else if len(peerAddrs) > 0 {
		peer := peerAddrs[0]
		ms.incomingInfoHashes[infoHash] = peerAddrs[1:]

		go NewClient(infoHash, &peer, ms.PeerID, ClientEventHandlers{
			OnSuccess: ms.flush,
			OnError:   ms.onLeechError,
		}).Do(time.Now().Add(ms.deadline))
	}
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

	var infoHash [20]byte
	copy(infoHash[:], result.InfoHash)
	delete(ms.incomingInfoHashes, infoHash)
}

func (ms *Sink) onLeechError(infoHash [20]byte, err error) {
	log.Debug().Err(err)

	ms.incomingInfoHashesMx.Lock()
	defer ms.incomingInfoHashesMx.Unlock()

	if len(ms.incomingInfoHashes[infoHash]) > 0 {
		peer := ms.incomingInfoHashes[infoHash][0]
		ms.incomingInfoHashes[infoHash] = ms.incomingInfoHashes[infoHash][1:]
		go NewClient(infoHash, &peer, ms.PeerID, ClientEventHandlers{
			OnSuccess: ms.flush,
			OnError:   ms.onLeechError,
		}).Do(time.Now().Add(ms.deadline))
	} else {
		ms.deleted++
		delete(ms.incomingInfoHashes, infoHash)
	}
}
