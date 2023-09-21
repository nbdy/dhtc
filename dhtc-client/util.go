package dhtc_client

import (
	"math/rand"
)

func randomDigit() byte {
	var max, min int
	max, min = '9', '0'
	return byte(rand.Intn(max-min) + min)
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
