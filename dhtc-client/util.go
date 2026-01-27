package dhtc_client

import (
	"encoding/binary"
	"math/rand"
	"net"
)

func randomDigit() byte {
	var high, low int
	high, low = '9', '0'
	return byte(rand.Intn(high-low) + low) //nolint:gosec // weak random number generator is fine for this use case
}

func randomID() []byte {
	/* > The peer_id is exactly 20 bytes (characters) long.
	 * >
	 * > There are mainly two conventions how to encode client and client version information into the peer_id,
	 * > Azureus-style and Shadow's-style.
	 * >
	 * > Azureus-style uses the following encoding: '-', two characters for client id, four ascii digits for version
	 * > number, '-', followed by random numbers.
	 * >
	 * > For example: '-AZ2060-'...
	 *
	 * https://wiki.theory.org/index.php/BitTorrentSpecification
	 *
	 * We encode the version number as:
	 *  - First two digits for the major version number
	 *  - Last two digits for the minor version number
	 *  - Patch version number is not encoded.
	 */
	prefix := []byte("-MC0008-")

	var rando []byte
	for i := 20 - len(prefix); i >= 0; i-- {
		rando = append(rando, randomDigit())
	}

	return append(prefix, rando...)
}

func parsePeers(s string) []net.TCPAddr {
	var peers []net.TCPAddr
	for i := 0; i+6 <= len(s); i += 6 {
		ip := net.IP(s[i : i+4])
		port := binary.BigEndian.Uint16([]byte(s[i+4 : i+6]))
		peers = append(peers, net.TCPAddr{IP: ip, Port: int(port)})
	}
	return peers
}

func parsePeers6(s string) []net.TCPAddr {
	var peers []net.TCPAddr
	for i := 0; i+18 <= len(s); i += 18 {
		ip := net.IP(s[i : i+16])
		port := binary.BigEndian.Uint16([]byte(s[i+16 : i+18]))
		peers = append(peers, net.TCPAddr{IP: ip, Port: int(port)})
	}
	return peers
}
