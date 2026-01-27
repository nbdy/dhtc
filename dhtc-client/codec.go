// Package dhtc_client provides a DHT client implementation.
package dhtc_client

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/anacrolix/missinggo/v2/iter"
	"github.com/anacrolix/torrent/bencode"
	"github.com/willf/bloom"
)

// Message represents a KRPC message.
type Message struct {
	// Q is the Query method. One of 5:
	//   - "ping"
	//   - "find_node"
	//   - "get_peers"
	//   - "announce_peer"
	//   - "sample_infohashes" (added by BEP 51)
	Q string `bencode:"q,omitempty"`
	// A contains the named QueryArguments sent with a query.
	A QueryArguments `bencode:"a,omitempty"`
	// T is the required transaction ID.
	T []byte `bencode:"t"`
	// Y is the required type of the message: q for QUERY, r for RESPONSE, e for ERROR.
	Y string `bencode:"y"`
	// R is the RESPONSE type only.
	R ResponseValues `bencode:"r,omitempty"`
	// E is the ERROR type only.
	E Error `bencode:"e,omitempty"`
}

// QueryArguments represents the "a" dictionary in a DHT query.
type QueryArguments struct {
	// ID is the ID of the querying Node.
	ID []byte `bencode:"id"`
	// InfoHash is the InfoHash of the torrent.
	InfoHash []byte `bencode:"info_hash,omitempty"`
	// Target is the ID of the node sought.
	Target []byte `bencode:"target,omitempty"`
	// Token is the token received from an earlier get_peers query.
	Token []byte `bencode:"token,omitempty"`
	// Port is the senders torrent port.
	Port int `bencode:"port,omitempty"`
	// ImpliedPort indicates if the senders apparent DHT port should be used.
	ImpliedPort int `bencode:"implied_port,omitempty"`

	// Seed indicates whether the querying node is seeding the torrent it announces.
	// Defined in BEP 33 "DHT Scrapes" for `announce_peer` queries.
	Seed int `bencode:"seed,omitempty"`

	// NoSeed indicates if the responding node should try to fill the `values` list with non-seed items.
	// Defined in BEP 33 "DHT Scrapes" for `get_peers` queries.
	NoSeed int `bencode:"noseed,omitempty"`
	// Scrape indicates if the responding node should add Bloom Filters to the response.
	// Defined in BEP 33 "DHT Scrapes" for `get_peers` queries.
	Scrape int `bencode:"scrape,omitempty"`
}

// ResponseValues represents the "r" dictionary in a DHT response.
type ResponseValues struct {
	// ID of the responding node.
	ID []byte `bencode:"id"`
	// Nodes is a list of K closest nodes to the requested target (IPv4).
	Nodes CompactNodeInfos `bencode:"nodes,omitempty"`
	// Nodes6 is a list of K closest nodes to the requested target (IPv6).
	Nodes6 CompactNodeInfos `bencode:"nodes6,omitempty"`
	// Token for future announce_peer.
	Token []byte `bencode:"token,omitempty"`
	// Values is a list of torrent peers.
	Values CompactPeers `bencode:"values,omitempty"`

	// Interval is the subset refresh interval in seconds (BEP 51).
	Interval int `bencode:"interval,omitempty"`
	// Num is the number of infohashes in storage (BEP 51).
	Num int `bencode:"num,omitempty"`
	// Samples is a subset of stored infohashes, N × 20 bytes (BEP 51).
	Samples []byte `bencode:"samples,omitempty"`
	// Samples2 is a subset of stored 32-byte infohashes (BEP 52).
	Samples2 [][]byte `bencode:"samples2,omitempty"`

	// BFsd is a Bloom Filter (256 bytes) representing all stored seeds for that infohash (BEP 33).
	BFsd *bloom.BloomFilter `bencode:"BFsd,omitempty"`
	// BFpe is a Bloom Filter (256 bytes) representing all stored peers for that infohash (BEP 33).
	BFpe *bloom.BloomFilter `bencode:"BFpe,omitempty"`
}

// Error represents a KRPC error.
type Error struct {
	Code    int
	Message []byte
}

// CompactPeer represents a peer's IP and port.
type CompactPeer struct {
	IP   net.IP
	Port int
}

// CompactPeers is a slice of CompactPeer.
type CompactPeers []CompactPeer

// CompactNodeInfo represents a node's ID and address.
type CompactNodeInfo struct {
	ID   []byte
	Addr net.UDPAddr
}

// CompactNodeInfos is a slice of CompactNodeInfo.
type CompactNodeInfos []CompactNodeInfo

// UnmarshalBencode unmarshals the compact peers from bencode.
// It supports both a list of strings and a single string.
func (cps *CompactPeers) UnmarshalBencode(b []byte) error {
	var list [][]byte
	if err := bencode.Unmarshal(b, &list); err == nil {
		*cps = make(CompactPeers, 0, len(list))
		for _, s := range list {
			var cp CompactPeer
			if err := cp.UnmarshalBinary(s); err != nil {
				return err
			}
			*cps = append(*cps, cp)
		}
		return nil
	}
	var bb []byte
	if err := bencode.Unmarshal(b, &bb); err != nil {
		return err
	}
	var err error
	*cps, err = UnmarshalCompactPeers(bb)
	return err
}

// MarshalBencode marshals the compact peers to bencode.
func (cps *CompactPeers) MarshalBencode() ([]byte, error) {
	list := make([][]byte, 0, len(*cps))
	for _, cp := range *cps {
		list = append(list, cp.MarshalBinary())
	}
	return bencode.Marshal(list)
}

// MarshalBinary marshals the compact peer to binary.
func (cp *CompactPeer) MarshalBinary() []byte {
	ip := cp.IP.To4()
	if ip == nil {
		ip = cp.IP.To16()
	}
	if ip == nil {
		return nil
	}
	ret := make([]byte, len(ip)+2)
	copy(ret, ip)
	binary.BigEndian.PutUint16(ret[len(ip):], uint16(cp.Port))
	return ret
}

// UnmarshalBinary unmarshals the compact peer from binary.
func (cp *CompactPeer) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case 18:
		cp.IP = make([]byte, 16)
	case 6:
		cp.IP = make([]byte, 4)
	default:
		return fmt.Errorf("bad compact peer string: %q", b)
	}
	copy(cp.IP, b)
	b = b[len(cp.IP):]
	cp.Port = int(binary.BigEndian.Uint16(b))
	return nil
}

// UnmarshalCompactPeers unmarshals the compact peers from a byte slice.
func UnmarshalCompactPeers(b []byte) (ret CompactPeers, err error) {
	if len(b) == 0 {
		return nil, nil
	}

	var peerSize int
	if len(b)%6 == 0 {
		peerSize = 6
	} else if len(b)%18 == 0 {
		peerSize = 18
	} else {
		return nil, fmt.Errorf("compact peer info length %d is neither a multiple of 6 nor 18", len(b))
	}

	num := len(b) / peerSize
	ret = make(CompactPeers, num)
	for i := range iter.N(num) {
		off := i * peerSize
		err = ret[i].UnmarshalBinary(b[off : off+peerSize])
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalBencode unmarshals the compact node infos from bencode.
func (cnis *CompactNodeInfos) UnmarshalBencode(b []byte) (err error) {
	var bb []byte
	err = bencode.Unmarshal(b, &bb)
	if err != nil {
		return
	}
	*cnis, err = UnmarshalCompactNodeInfos(bb)
	return
}

// UnmarshalCompactNodeInfos unmarshals the compact node infos from a byte slice.
func UnmarshalCompactNodeInfos(b []byte) (ret []CompactNodeInfo, err error) {
	if len(b) == 0 {
		return nil, nil
	}

	var nodeSize int
	if len(b)%26 == 0 {
		nodeSize = 26
	} else if len(b)%38 == 0 {
		nodeSize = 38
	} else {
		err = fmt.Errorf("compact node is neither a multiple of 26 nor 38")
		return
	}

	num := len(b) / nodeSize
	ret = make([]CompactNodeInfo, num)
	for i := range iter.N(num) {
		off := i * nodeSize
		err = ret[i].UnmarshalBinary(b[off : off+nodeSize])
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalBinary unmarshals the compact node info from binary.
func (cni *CompactNodeInfo) UnmarshalBinary(b []byte) error {
	if len(b) != 26 && len(b) != 38 {
		return fmt.Errorf("invalid compact node info length: %d", len(b))
	}
	if len(cni.ID) != 20 {
		cni.ID = make([]byte, 20)
	}
	copy(cni.ID, b[:20])
	b = b[20:]

	var ipLen int
	if len(b) == 6 {
		ipLen = 4
	} else {
		ipLen = 16
	}

	cni.Addr.IP = make([]byte, ipLen)
	copy(cni.Addr.IP, b[:ipLen])
	b = b[ipLen:]
	cni.Addr.Port = int(binary.BigEndian.Uint16(b))
	cni.Addr.Zone = ""
	return nil
}

// MarshalBencode marshals the compact node infos to bencode.
func (cnis *CompactNodeInfos) MarshalBencode() ([]byte, error) {
	var ret []byte

	if len(*cnis) == 0 {
		return []byte("0:"), nil
	}

	for _, cni := range *cnis {
		ret = append(ret, cni.MarshalBinary()...)
	}

	return bencode.Marshal(ret)
}

// MarshalBinary marshals the compact node info to binary.
func (cni *CompactNodeInfo) MarshalBinary() []byte {
	ret := make([]byte, 20)
	copy(ret, cni.ID)

	ip := cni.Addr.IP
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}
	ret = append(ret, ip...)

	portEncoding := make([]byte, 2)
	binary.BigEndian.PutUint16(portEncoding, uint16(cni.Addr.Port))
	ret = append(ret, portEncoding...)

	return ret
}

// MarshalBencode marshals the error to bencode.
func (e *Error) MarshalBencode() ([]byte, error) {
	return bencode.Marshal([]interface{}{e.Code, e.Message})
}

// UnmarshalBencode unmarshals the error from bencode.
func (e *Error) UnmarshalBencode(b []byte) (err error) {
	var i interface{}
	err = bencode.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	l, ok := i.([]interface{})
	if !ok {
		return fmt.Errorf("invalid error type: %T", i)
	}

	if len(l) < 2 {
		return fmt.Errorf("invalid error list length: %d", len(l))
	}

	code, ok := l[0].(int64)
	if !ok {
		return fmt.Errorf("invalid error code type: %T", l[0])
	}
	e.Code = int(code)

	switch v := l[1].(type) {
	case []byte:
		e.Message = v
	case string:
		e.Message = []byte(v)
	default:
		return fmt.Errorf("invalid error message type: %T", l[1])
	}

	return nil
}
