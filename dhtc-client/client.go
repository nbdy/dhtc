package dhtc_client

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const MaxMetadataSize = 10 * 1024 * 1024

type Client struct {
	infoHash []byte
	peerAddr *net.TCPAddr
	ev       ClientEventHandlers

	conn     *net.TCPConn
	clientID [20]byte

	utMetadata                     uint8
	utPex                          uint8
	metadataReceived, metadataSize uint
	metadata                       []byte

	connClosed bool
}

type ClientEventHandlers struct {
	OnSuccess func(Metadata)              // must be supplied. args: metadata
	OnError   func([]byte, error)         // must be supplied. args: infohash, error
	OnPeers   func([]byte, []net.TCPAddr) // args: infohash, peers
}

func NewClient(infoHash []byte, peerAddr *net.TCPAddr, clientID []byte, ev ClientEventHandlers) *Client {
	l := new(Client)
	l.infoHash = infoHash
	l.peerAddr = peerAddr
	copy(l.clientID[:], clientID)
	l.ev = ev
	return l
}

func (c *Client) writeAll(b []byte) error {
	for len(b) != 0 {
		n, err := c.conn.Write(b)
		if err != nil {
			return err
		}
		b = b[n:]
	}
	return nil
}

func (c *Client) doBtHandshake() error {
	ih := c.infoHash
	if len(ih) == 32 {
		ih = ih[:20]
	}

	lHandshake := make([]byte, 68)
	lHandshake[0] = 19
	copy(lHandshake[1:20], "BitTorrent protocol")
	lHandshake[25] = 0x10
	lHandshake[27] = 0x01
	copy(lHandshake[28:48], ih)
	copy(lHandshake[48:68], c.clientID[:])

	err := c.writeAll(lHandshake)
	if err != nil {
		return errors.Wrap(err, "writeAll lHandshake")
	}

	rHandshake, err := c.readExactly(68)
	if err != nil {
		return errors.Wrap(err, "readExactly rHandshake")
	}

	if !bytes.HasPrefix(rHandshake, []byte("\x13BitTorrent protocol")) {
		return fmt.Errorf("corrupt BitTorrent handshake received")
	}

	if !bytes.Equal(rHandshake[28:48], ih) {
		return fmt.Errorf("remote peer infohash mismatch")
	}

	if (rHandshake[25] & 0x10) == 0 {
		return fmt.Errorf("peer does not support the extension protocol")
	}

	return nil
}

func (c *Client) doExHandshake() error {
	err := c.writeAll([]byte("\x00\x00\x00\x25\x14\x00d1:md11:ut_metadatai1e6:ut_pexi2eee"))
	if err != nil {
		return errors.Wrap(err, "writeAll lHandshake")
	}

	rExMessage, err := c.readExMessage()
	if err != nil {
		return errors.Wrap(err, "readExMessage")
	}

	// Extension Handshake has the Extension Message ID = 0x00
	if rExMessage[1] != 0 {
		return errors.Wrap(err, "first extension message is not an extension handshake")
	}

	rRootDict := new(rootDict)
	err = bencode.Unmarshal(rExMessage[2:], rRootDict)
	if err != nil {
		return errors.Wrap(err, "unmarshal rExMessage")
	}

	if !(0 < rRootDict.MetadataSize && rRootDict.MetadataSize < MaxMetadataSize) {
		return fmt.Errorf("metadata too big or its size is less than or equal zero")
	}

	if !(0 < rRootDict.M.UTMetadata && rRootDict.M.UTMetadata < 255) {
		return fmt.Errorf("ut_metadata is not an uint8")
	}

	c.utMetadata = uint8(rRootDict.M.UTMetadata) // Save the ut_metadata code the remote peer uses
	c.utPex = uint8(rRootDict.M.UTPex)
	c.metadataSize = uint(rRootDict.MetadataSize)
	c.metadata = make([]byte, c.metadataSize)

	return nil
}

func (c *Client) handlePex(payload []byte) {
	var msg pexMsg
	err := bencode.Unmarshal(payload, &msg)
	if err != nil {
		return
	}

	peers := parsePeers(msg.Added)
	peers = append(peers, parsePeers6(msg.Added6)...)

	if len(peers) > 0 && c.ev.OnPeers != nil {
		c.ev.OnPeers(c.infoHash, peers)
	}
}

func (c *Client) requestAllPieces() error {
	// Request all the pieces of metadata
	nPieces := int(math.Ceil(float64(c.metadataSize) / math.Pow(2, 14)))
	for piece := 0; piece < nPieces; piece++ {
		// __request_metadata_piece(piece)
		// ...............................
		extDictDump, err := bencode.Marshal(extDict{
			MsgType: 0,
			Piece:   piece,
		})
		if err != nil { // ASSERT
			panic(errors.Wrap(err, "marshal extDict"))
		}

		err = c.writeAll([]byte(fmt.Sprintf(
			"%s\x14%s%s",
			toBigEndian(uint(2+len(extDictDump)), 4),
			toBigEndian(uint(c.utMetadata), 1),
			extDictDump,
		)))
		if err != nil {
			return errors.Wrap(err, "writeAll piece request")
		}
	}

	return nil
}

// readMessage returns a BitTorrent message, sans the first 4 bytes indicating its length.
func (c *Client) readMessage() ([]byte, error) {
	rLengthB, err := c.readExactly(4)
	if err != nil {
		return nil, errors.Wrap(err, "readExactly rLengthB")
	}

	rLength := uint(binary.BigEndian.Uint32(rLengthB))

	// Some malicious/faulty peers say that they are sending a very long
	// message, and hence causing us to run out of memory.
	// This is a crude check that does not let it happen (i.e. boundary can probably be
	// tightened a lot more.)
	if rLength > MaxMetadataSize {
		return nil, errors.New("message is longer than max allowed metadata size")
	}

	rMessage, err := c.readExactly(rLength)
	if err != nil {
		return nil, errors.Wrap(err, "readExactly rMessage")
	}

	return rMessage, nil
}

func (c *Client) readExMessage() ([]byte, error) {
	for {
		rMessage, err := c.readMessage()
		if err != nil {
			return nil, errors.Wrap(err, "readMessage")
		}

		// Every extension message has at least 2 bytes.
		if len(rMessage) < 2 {
			continue
		}

		// We are interested only in extension messages, whose first byte is always 20
		if rMessage[0] == 20 {
			return rMessage, nil
		}
	}
}

func (c *Client) connect(deadline time.Time) error {
	var err error

	x, err := net.DialTimeout("tcp4", c.peerAddr.String(), 1*time.Second)
	if err != nil {
		return errors.Wrap(err, "dial")
	}
	c.conn = x.(*net.TCPConn)

	// > If sec == 0, operating system discards any unsent or unacknowledged data [after Close()
	// > has been called].
	err = c.conn.SetLinger(0)
	if err != nil {
		if err := c.conn.Close(); err != nil {
			log.Panic().Msg("couldn't close leech connection!")
			log.Panic().Err(err)
		}
		return errors.Wrap(err, "SetLinger")
	}

	err = c.conn.SetNoDelay(true)
	if err != nil {
		if err := c.conn.Close(); err != nil {
			log.Panic().Msg("couldn't close leech connection!")
			log.Panic().Err(err)
		}
		return errors.Wrap(err, "NODELAY")
	}

	err = c.conn.SetDeadline(deadline)
	if err != nil {
		if err := c.conn.Close(); err != nil {
			log.Panic().Msg("couldn't close leech connection!")
			log.Panic().Err(err)
		}
		return errors.Wrap(err, "SetDeadline")
	}

	return nil
}

func (c *Client) closeConn() {
	if c.connClosed {
		return
	}

	if err := c.conn.Close(); err != nil {
		log.Panic().Msg("couldn't close leech connection!")
		log.Panic().Err(err)
		return
	}

	c.connClosed = true
}

func (c *Client) Do(deadline time.Time) {
	err := c.connect(deadline)
	if err != nil {
		c.OnError(errors.Wrap(err, "connect"))
		return
	}
	defer c.closeConn()

	err = c.doBtHandshake()
	if err != nil {
		c.OnError(errors.Wrap(err, "doBtHandshake"))
		return
	}

	err = c.doExHandshake()
	if err != nil {
		c.OnError(errors.Wrap(err, "doExHandshake"))
		return
	}

	err = c.requestAllPieces()
	if err != nil {
		c.OnError(errors.Wrap(err, "requestAllPieces"))
		return
	}

	for c.metadataReceived < c.metadataSize {
		rExMessage, err := c.readExMessage()
		if err != nil {
			c.OnError(errors.Wrap(err, "readExMessage"))
			return
		}

		if rExMessage[1] == 1 { // ut_metadata
			rMessageBuf := bytes.NewBuffer(rExMessage[2:])
			rExtDict := new(extDict)
			err = bencode.NewDecoder(rMessageBuf).Decode(rExtDict)
			if err != nil {
				c.OnError(errors.Wrap(err, "could not decode ext msg in the loop"))
				return
			}

			if rExtDict.MsgType == 2 { // reject
				c.OnError(fmt.Errorf("remote peer rejected sending metadata"))
				return
			}

			if rExtDict.MsgType == 1 { // data
				// Get the unread bytes!
				metadataPiece := rMessageBuf.Bytes()

				// BEP 9 explicitly states:
				//   > If the piece is the last piece of the metadata, it may be less than 16kiB. If
				//   > it is not the last piece of the metadata, it MUST be 16kiB.
				//
				// Hence...
				//   ... if the length of @metadataPiece is more than 16kiB, we err.
				if len(metadataPiece) > 16*1024 {
					c.OnError(fmt.Errorf("metadataPiece > 16kiB"))
					return
				}

				piece := rExtDict.Piece
				// metadata[piece * 2**14: piece * 2**14 + len(metadataPiece)] = metadataPiece is how it'd be done in Python
				copy(c.metadata[piece*16384:piece*16384+len(metadataPiece)], metadataPiece)
				c.metadataReceived += uint(len(metadataPiece))

				// ... if the length of @metadataPiece is less than 16kiB AND metadata is NOT
				// complete then we err.
				if len(metadataPiece) < 16*1024 && c.metadataReceived != c.metadataSize {
					c.OnError(fmt.Errorf("metadataPiece < 16 kiB but incomplete"))
					return
				}

				if c.metadataReceived > c.metadataSize {
					c.OnError(fmt.Errorf("metadataReceived > metadataSize"))
					return
				}
			}
		} else if rExMessage[1] == 2 { // ut_pex
			c.handlePex(rExMessage[2:])
		}
	}

	// We are done with the transfer, close socket as soon as possible (i.e. NOW) to avoid hitting "too many open files"
	// error.
	c.closeConn()

	// Verify the checksum
	var sum []byte
	if len(c.infoHash) == 32 {
		s256 := sha256.Sum256(c.metadata)
		sum = s256[:]
	} else {
		s1 := sha1.Sum(c.metadata)
		sum = s1[:]
	}

	if !bytes.Equal(sum, c.infoHash) {
		c.OnError(fmt.Errorf("infohash mismatch"))
		return
	}

	// Check the info dictionary
	info := new(metainfo.Info)
	err = bencode.Unmarshal(c.metadata, info)
	if err != nil {
		c.OnError(errors.Wrap(err, "unmarshal info"))
		return
	}
	err = validateInfo(info)
	if err != nil {
		c.OnError(errors.Wrap(err, "validateInfo"))
		return
	}

	var files []File
	// If there is only one file, there won't be a Files slice. That's why we need to add it here
	if len(info.Files) == 0 {
		files = append(files, File{
			Size: info.Length,
			Path: info.Name,
		})
	} else {
		for _, file := range info.Files {
			files = append(files, File{
				Size: file.Length,
				Path: file.DisplayPath(info),
			})
		}
	}

	var totalSize uint64
	for _, file := range files {
		if file.Size < 0 {
			c.OnError(fmt.Errorf("file size less than zero"))
			return
		}

		totalSize += uint64(file.Size)
	}

	c.ev.OnSuccess(Metadata{
		InfoHash:     c.infoHash[:],
		Name:         info.Name,
		TotalSize:    totalSize,
		DiscoveredOn: time.Now().Unix(),
		Files:        files,
	})
}

func validateInfo(info *metainfo.Info) error {
	if len(info.Pieces) > 0 && len(info.Pieces)%20 != 0 {
		return errors.New("pieces has invalid length")
	}
	if info.PieceLength == 0 {
		if info.TotalLength() != 0 {
			return errors.New("zero piece length")
		}
	} else {
		if len(info.Pieces) > 0 && int((info.TotalLength()+info.PieceLength-1)/info.PieceLength) != info.NumPieces() {
			return errors.New("piece count and file lengths are at odds")
		}
	}
	return nil
}

func (c *Client) readExactly(n uint) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(c.conn, b)
	return b, err
}

func (c *Client) OnError(err error) {
	c.ev.OnError(c.infoHash, err)
}

func toBigEndian(i uint, n int) []byte {
	b := make([]byte, n)
	switch n {
	case 1:
		b = []byte{byte(i)}

	case 2:
		binary.BigEndian.PutUint16(b, uint16(i))

	case 4:
		binary.BigEndian.PutUint32(b, uint32(i))

	default:
		panic("n must be 1, 2, or 4!")
	}

	if len(b) != n {
		panic(fmt.Sprintf("postcondition failed: len(b) != n in intToBigEndian (i %d, n %d, len b %d, b %s)", i, n, len(b), b))
	}

	return b
}
