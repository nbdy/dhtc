package dhtc_client

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/rs/zerolog/log"
	"net"
)

type Transport struct {
	fd      *net.UDPConn
	laddr   *net.UDPAddr
	started bool
	buffer  []byte

	// OnMessage is the function that will be called when Transport receives a packet that is
	// successfully unmarshalled as a syntactically correct Message (but -of course- the checking
	// the semantic correctness of the Message is left to Protocol).
	onMessage func(*Message, *net.UDPAddr)
	// OnCongestion
	onCongestion func()
}

func NewTransport(laddr string, onMessage func(*Message, *net.UDPAddr), onCongestion func()) *Transport {
	t := new(Transport)
	/*   The field size sets a theoretical limit of 65,535 bytes (8 byte header + 65,527 bytes of
	 * data) for a UDP datagram. However the actual limit for the data length, which is imposed by
	 * the underlying IPv4 protocol, is 65,507 bytes (65,535 − 8 byte UDP header − 20 byte IP
	 * header).
	 *
	 *   In IPv6 jumbograms it is possible to have UDP packets of size greater than 65,535 bytes.
	 * RFC 2675 specifies that the length field is set to zero if the length of the UDP header plus
	 * UDP data is greater than 65,535.
	 *
	 * https://en.wikipedia.org/wiki/User_Datagram_Protocol
	 */
	t.buffer = make([]byte, 65507)
	t.onMessage = onMessage
	t.onCongestion = onCongestion

	var err error
	t.laddr, err = net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		log.Panic().Msg("Could not resolve the UDP address for the trawler!")
		log.Panic().Err(err)
	}
	if t.laddr.IP.To4() == nil {
		log.Panic().Msg("IP address is not IPv4!")
	}

	return t
}

func (t *Transport) Start() {
	// Why check whether the Transport `t` started or not, here and not -for instance- in
	// t.Terminate()?
	// Because in t.Terminate() the programmer (i.e. you & me) would stumble upon an error while
	// trying close an uninitialised net.UDPConn or something like that: it's mostly harmless
	// because its effects are immediate. But if you try to start a Transport `t` for the second
	// (or the third, 4th, ...) time, it will keep spawning goroutines and any small mistake may
	// end up in a debugging horror.
	//                                                                   Here ends my justification.
	if t.started {
		log.Panic().Msg("Attempting to Start() a mainline/Transport that has been already started! (Programmer error.)")
	}
	t.started = true

	var err error
	var ip [4]byte
	copy(ip[:], t.laddr.IP.To4())

	t.fd, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   t.laddr.IP.To4(),
		Port: t.laddr.Port,
	})

	if err != nil {
		log.Fatal().Msg("Could NOT bind the socket!")
		log.Fatal().Err(err)
	}

	go t.readMessages()
}

func (t *Transport) Terminate() {
	_ = t.fd.Close()
}

// readMessages is a goroutine!
func (t *Transport) readMessages() {
	for {
		n, fromSA, err := t.fd.ReadFromUDP(t.buffer)

		if n == 0 {
			/* Datagram sockets in various domains  (e.g., the UNIX and Internet domains) permit
			 * zero-length datagrams. When such a datagram is received, the return value (n) is 0.
			 */
			continue
		}

		var msg Message
		err = bencode.Unmarshal(t.buffer[:n], &msg)
		if err != nil {
			// couldn't unmarshal packet data
			continue
		}

		t.onMessage(&msg, fromSA)
	}
}

func (t *Transport) WriteMessages(msg *Message, addr *net.UDPAddr) {
	data, err := bencode.Marshal(msg)
	if err != nil {
		log.Panic().Msg("Could NOT marshal an outgoing message! (Programmer error.)")
	}

	_, err = t.fd.WriteToUDP(data, addr)
	if err != nil {
		log.Warn().Msg("Could NOT write an UDP packet!")
		log.Warn().Err(err)
	}
}
