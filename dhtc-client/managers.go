package dhtc_client

import (
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type Service interface {
	Start(nodes []string)
	Terminate()
}

type Result interface {
	InfoHash() [20]byte
	PeerAddrs() []net.TCPAddr
}

type Manager struct {
	output           chan Result
	indexingServices []Service
}

func NewManager(nodes []string, addrs []string, interval time.Duration, maxNeighbors uint) *Manager {
	manager := new(Manager)
	manager.output = make(chan Result, 20)

	for _, addr := range addrs {
		service := NewIndexingService(addr, interval, maxNeighbors, IndexingServiceEventHandlers{
			OnResult: manager.onIndexingResult,
		})
		manager.indexingServices = append(manager.indexingServices, service)
		service.Start(nodes)
	}

	return manager
}

func (m *Manager) onIndexingResult(res IndexingResult) {
	select {
	case m.output <- res:
	default:
		log.Debug().Msg("DHT manager output ch is full, idx result dropped!")
	}
}

func (m *Manager) Output() <-chan Result {
	return m.output
}

func (m *Manager) Terminate() {
	for _, service := range m.indexingServices {
		service.Terminate()
	}
}
