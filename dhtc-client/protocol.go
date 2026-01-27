package dhtc_client

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Protocol represents a DHT protocol handler.
type Protocol struct {
	previousTokenSecret, currentTokenSecret []byte
	tokenLock                               sync.Mutex
	transport                               *Transport
	eventHandlers                           ProtocolEventHandlers
	started                                 bool
}

// ProtocolEventHandlers contains the callback functions for various DHT events.
type ProtocolEventHandlers struct {
	// OnPingQuery is called when a ping query is received.
	OnPingQuery func(*Message, *net.UDPAddr)
	// OnFindNodeQuery is called when a find_node query is received.
	OnFindNodeQuery func(*Message, *net.UDPAddr)
	// OnGetPeersQuery is called when a get_peers query is received.
	OnGetPeersQuery func(*Message, *net.UDPAddr)
	// OnAnnouncePeerQuery is called when an announce_peer query is received.
	OnAnnouncePeerQuery func(*Message, *net.UDPAddr)
	// OnGetPeersResponse is called when a get_peers response is received.
	OnGetPeersResponse func(*Message, *net.UDPAddr)
	// OnFindNodeResponse is called when a find_node response is received.
	OnFindNodeResponse func(*Message, *net.UDPAddr)
	// OnPingORAnnouncePeerResponse is called when a ping or announce_peer response is received.
	OnPingORAnnouncePeerResponse func(*Message, *net.UDPAddr)

	// OnSampleInfohashesQuery is called when a sample_infohashes query is received (BEP 51).
	OnSampleInfohashesQuery func(*Message, *net.UDPAddr)
	// OnSampleInfohashesResponse is called when a sample_infohashes response is received (BEP 51).
	OnSampleInfohashesResponse func(*Message, *net.UDPAddr)

	// OnCongestion is called when congestion is detected.
	OnCongestion func()
}

// NewProtocol creates a new DHT protocol handler.
func NewProtocol(laddr string, rateLimit int, eventHandlers ProtocolEventHandlers) (p *Protocol) {
	p = new(Protocol)
	p.eventHandlers = eventHandlers
	p.transport = NewTransport(laddr, rateLimit, p.onMessage, p.eventHandlers.OnCongestion)

	p.currentTokenSecret, p.previousTokenSecret = make([]byte, 20), make([]byte, 20)
	_, err := rand.Read(p.currentTokenSecret)
	if err != nil {
		log.Fatal().Msg("Could NOT generate random bytes for token secret!")
		log.Fatal().Err(err)
	}
	copy(p.previousTokenSecret, p.currentTokenSecret)

	return
}

// Start starts the DHT protocol handler.
func (p *Protocol) Start() {
	if p.started {
		log.Panic().Msg("Attempting to Start() a mainline/Protocol that has been already started! (Programmer error.)")
	}
	p.started = true

	p.transport.Start()
	go p.updateTokenSecret()
}

// Terminate terminates the DHT protocol handler.
func (p *Protocol) Terminate() {
	if !p.started {
		log.Panic().Msg("Attempted to Terminate() a mainline/Protocol that has not been Start()ed! (Programmer error.)")
	}

	p.transport.Terminate()
}

func (p *Protocol) onMessage(msg *Message, addr *net.UDPAddr) {
	switch msg.Y {
	case "q":
		switch msg.Q {
		case "ping":
			if !validatePingQueryMessage(msg) {
				// zap.L().Debug("An invalid ping query received!")
				return
			}
			// Check whether there is a registered event handler for the ping queries, before
			// attempting to call.
			if p.eventHandlers.OnPingQuery != nil {
				p.eventHandlers.OnPingQuery(msg, addr)
			}

		case "find_node":
			if !validateFindNodeQueryMessage(msg) {
				// zap.L().Debug("An invalid find_node query received!")
				return
			}
			if p.eventHandlers.OnFindNodeQuery != nil {
				p.eventHandlers.OnFindNodeQuery(msg, addr)
			}

		case "get_peers":
			if !validateGetPeersQueryMessage(msg) {
				// zap.L().Debug("An invalid get_peers query received!")
				return
			}
			if p.eventHandlers.OnGetPeersQuery != nil {
				p.eventHandlers.OnGetPeersQuery(msg, addr)
			}

		case "announce_peer":
			if !validateAnnouncePeerQueryMessage(msg) {
				// zap.L().Debug("An invalid announce_peer query received!")
				return
			}
			if p.eventHandlers.OnAnnouncePeerQuery != nil {
				p.eventHandlers.OnAnnouncePeerQuery(msg, addr)
			}

		case "vote":
			// Although we are aware that such method exists, we ignore.

		case "sample_infohashes": // Added by BEP 51
			if !validateSampleInfohashesQueryMessage(msg) {
				// zap.L().Debug("An invalid sample_infohashes query received!")
				return
			}
			if p.eventHandlers.OnSampleInfohashesQuery != nil {
				p.eventHandlers.OnSampleInfohashesQuery(msg, addr)
			}

		default:
			// zap.L().Debug("A KRPC query of an unknown method received!", zap.String("method", msg.Q))
			return
		}
	case "r":
		// Query messages have a `q` field which indicates their type but response messages have no such field that we
		// can rely on.
		// The idea is you'd use transaction ID (the `t` key) to deduce the type of response message, as it must be
		// sent in response to a query message (with the same transaction ID) that we have sent earlier.
		// This approach is, unfortunately, not very practical for our needs since we send up to thousands messages per
		// second, meaning that we'd run out of transaction IDs very quickly (since some [many?] clients assume
		// transaction IDs are no longer than 2 bytes), and we'd also then have to consider retention too (as we might
		// not get a response at all).
		// Our approach uses an ad-hoc pattern matching: all response messages share a subset of fields (such as `t`,
		// `y`) but only one type of them contain a particular field (such as `token` field is unique to `get_peers`
		// responses, `samples` is unique to `sample_infohashes` etc.).
		//
		// sample_infohashes > get_peers > find_node > ping / announce_peer
		if len(msg.R.Samples) != 0 { // The message should be a sample_infohashes response.
			if !validateSampleInfohashesResponseMessage(msg) {
				// zap.L().Debug("An invalid sample_infohashes response received!")
				return
			}
			if p.eventHandlers.OnSampleInfohashesResponse != nil {
				p.eventHandlers.OnSampleInfohashesResponse(msg, addr)
			}
		} else if len(msg.R.Token) != 0 { // The message should be a get_peers response.
			if !validateGetPeersResponseMessage(msg) {
				// zap.L().Debug("An invalid get_peers response received!")
				return
			}
			if p.eventHandlers.OnGetPeersResponse != nil {
				p.eventHandlers.OnGetPeersResponse(msg, addr)
			}
		} else if len(msg.R.Nodes) != 0 { // The message should be a find_node response.
			if !validateFindNodeResponseMessage(msg) {
				// zap.L().Debug("An invalid find_node response received!")
				return
			}
			if p.eventHandlers.OnFindNodeResponse != nil {
				p.eventHandlers.OnFindNodeResponse(msg, addr)
			}
		} else { // The message should be a ping or an announce_peer response.
			if !validatePingORannouncePeerResponseMessage(msg) {
				// zap.L().Debug("An invalid ping OR announce_peer response received!")
				return
			}
			if p.eventHandlers.OnPingORAnnouncePeerResponse != nil {
				p.eventHandlers.OnPingORAnnouncePeerResponse(msg, addr)
			}
		}
	case "e":
		// Ignore the following:
		//   - 202  Server Error
		//   - 204  Method Unknown / Unknown query type
		if msg.E.Code != 202 && msg.E.Code != 204 {
			log.Debug().Msgf("Protocol error received: `%s` (%d)", msg.E.Message, msg.E.Code)
		}
	default:
		/* zap.L().Debug("A KRPC message of an unknown type received!",
		zap.String("type", msg.Y))
		*/
	}
}

// SendMessage sends a KRPC message to the specified address.
func (p *Protocol) SendMessage(msg *Message, addr *net.UDPAddr) {
	p.transport.WriteMessages(msg, addr)
}

// NewFindNodeQuery creates a new find_node query message.
func NewFindNodeQuery(id []byte, target []byte) *Message {
	return &Message{
		Y: "q",
		T: []byte("aa"),
		Q: "find_node",
		A: QueryArguments{
			ID:     id,
			Target: target,
		},
	}
}

// NewGetPeersQuery creates a new get_peers query message.
func NewGetPeersQuery(id []byte, infoHash []byte) *Message {
	return &Message{
		Y: "q",
		T: []byte("aa"),
		Q: "get_peers",
		A: QueryArguments{
			ID:       id,
			InfoHash: infoHash,
		},
	}
}

// NewSampleInfohashesQuery creates a new sample_infohashes query message.
func NewSampleInfohashesQuery(id []byte, t []byte, target []byte) *Message {
	return &Message{
		Y: "q",
		T: t,
		Q: "sample_infohashes",
		A: QueryArguments{
			ID:     id,
			Target: target,
		},
	}
}

// NewSampleInfohashesResponse creates a new sample_infohashes response message.
func NewSampleInfohashesResponse(t []byte, id []byte, interval int, nodes CompactNodeInfos, nodes6 CompactNodeInfos, num int, samples []byte) *Message {
	return &Message{
		Y: "r",
		T: t,
		R: ResponseValues{
			ID:       id,
			Interval: interval,
			Nodes:    nodes,
			Nodes6:   nodes6,
			Num:      num,
			Samples:  samples,
		},
	}
}

// CalculateToken calculates a token for the specified address.
func (p *Protocol) CalculateToken(address net.IP) []byte {
	p.tokenLock.Lock()
	defer p.tokenLock.Unlock()
	sum := sha1.Sum(append(p.currentTokenSecret, address...))
	return sum[:]
}

// VerifyToken verifies the token for the specified address.
func (p *Protocol) VerifyToken(address net.IP, token []byte) bool {
	p.tokenLock.Lock()
	defer p.tokenLock.Unlock()

	sum := sha1.Sum(append(p.currentTokenSecret, address...))
	if bytes.Equal(sum[:], token) {
		return true
	}

	sum = sha1.Sum(append(p.previousTokenSecret, address...))
	return bytes.Equal(sum[:], token)
}

func (p *Protocol) updateTokenSecret() {
	for range time.Tick(10 * time.Minute) {
		p.tokenLock.Lock()
		copy(p.previousTokenSecret, p.currentTokenSecret)
		_, err := rand.Read(p.currentTokenSecret)
		if err != nil {
			p.tokenLock.Unlock()
			log.Fatal().Msg("Could NOT generate random bytes for token secret!")
			log.Fatal().Err(err)
		}
		p.tokenLock.Unlock()
	}
}

func validatePingQueryMessage(msg *Message) bool {
	return len(msg.A.ID) == 20
}

func validateFindNodeQueryMessage(msg *Message) bool {
	return len(msg.A.ID) == 20 &&
		len(msg.A.Target) == 20
}

func validateGetPeersQueryMessage(msg *Message) bool {
	return len(msg.A.ID) == 20 &&
		(len(msg.A.InfoHash) == 20 || len(msg.A.InfoHash) == 32)
}

func validateAnnouncePeerQueryMessage(msg *Message) bool {
	return len(msg.A.ID) == 20 &&
		(len(msg.A.InfoHash) == 20 || len(msg.A.InfoHash) == 32) &&
		msg.A.Port > 0 &&
		len(msg.A.Token) > 0
}

func validateSampleInfohashesQueryMessage(msg *Message) bool {
	return len(msg.A.ID) == 20 &&
		len(msg.A.Target) == 20
}

func validatePingORannouncePeerResponseMessage(msg *Message) bool {
	return len(msg.R.ID) == 20
}

func validateFindNodeResponseMessage(msg *Message) bool {
	if len(msg.R.ID) != 20 {
		return false
	}

	return len(msg.R.Nodes) > 0 || len(msg.R.Nodes6) > 0
}

func validateGetPeersResponseMessage(msg *Message) bool {
	if len(msg.R.ID) != 20 || len(msg.R.Token) == 0 {
		return false
	}

	return len(msg.R.Values) > 0 || len(msg.R.Nodes) > 0 || len(msg.R.Nodes6) > 0
}

func validateSampleInfohashesResponseMessage(msg *Message) bool {
	if len(msg.R.ID) != 20 || msg.R.Interval < 0 || msg.R.Num < 0 {
		return false
	}
	if len(msg.R.Samples)%20 != 0 {
		return false
	}
	for _, sample := range msg.R.Samples2 {
		if len(sample) != 32 {
			return false
		}
	}
	return true
}
